/*
 * Copyright (c) 2019 SUSE LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package kubernetes

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes"
)

var (
	// Container runtime is of the form `docker://version` or `cri-o://version`
	runtimeVersionRegexp = regexp.MustCompile(`//(.+)$`)
)

type NodeVersionInfo struct {
	Nodename                 string
	ContainerRuntimeVersion  *version.Version
	KubeletVersion           *version.Version
	APIServerVersion         *version.Version
	ControllerManagerVersion *version.Version
	SchedulerVersion         *version.Version
	EtcdVersion              *version.Version
	Unschedulable            bool
}

type NodeVersionInfoMap map[string]NodeVersionInfo

type VersionInquirer interface {
	AvailablePlatformVersions() []*version.Version
	NodeVersionInfoForClusterVersion(node *v1.Node, version *version.Version) NodeVersionInfo
}
type StaticVersionInquirer struct{}

func (si StaticVersionInquirer) AvailablePlatformVersions() []*version.Version {
	return AvailableVersionsForMap(Versions)
}

func (si StaticVersionInquirer) NodeVersionInfoForClusterVersion(node *v1.Node, clusterVersion *version.Version) NodeVersionInfo {
	res := NodeVersionInfo{
		Nodename:                node.ObjectMeta.Name,
		ContainerRuntimeVersion: version.MustParseSemantic(ComponentVersionForClusterVersion(ContainerRuntime, clusterVersion)),
		KubeletVersion:          version.MustParseSemantic(ComponentVersionForClusterVersion(Kubelet, clusterVersion)),
	}
	if IsControlPlane(node) {
		res.APIServerVersion = version.MustParseSemantic(ComponentVersionForClusterVersion(Hyperkube, clusterVersion))
		res.ControllerManagerVersion = version.MustParseSemantic(ComponentVersionForClusterVersion(Hyperkube, clusterVersion))
		res.SchedulerVersion = version.MustParseSemantic(ComponentVersionForClusterVersion(Hyperkube, clusterVersion))
		res.EtcdVersion = version.MustParseSemantic(ComponentVersionForClusterVersion(Etcd, clusterVersion))
	}
	return res
}

func (nvi NodeVersionInfo) IsControlPlane() bool {
	return nvi.APIServerVersion != nil
}

func (nvi NodeVersionInfo) String() string {
	if nvi.IsControlPlane() {
		return nvi.APIServerVersion.String()
	}
	return nvi.KubeletVersion.String()
}

func (nvi NodeVersionInfo) EqualsClusterVersion(clusterVersion *version.Version) bool {
	if nvi.IsControlPlane() {
		if nvi.APIServerVersion.String() != clusterVersion.String() {
			return false
		}
	}
	return nvi.KubeletVersion.String() == clusterVersion.String()
}

func (nvi NodeVersionInfo) LessThanClusterVersion(clusterVersion *version.Version) bool {
	if nvi.IsControlPlane() {
		if nvi.APIServerVersion.LessThan(clusterVersion) {
			return true
		}
	}
	return nvi.KubeletVersion.LessThan(clusterVersion)
}

func (nvi NodeVersionInfo) DriftsFromClusterVersion(clusterVersion *version.Version) bool {
	if nvi.IsControlPlane() {
		if nvi.APIServerVersion.Major() < clusterVersion.Major() ||
			nvi.APIServerVersion.Minor() < clusterVersion.Minor() {
			return true
		}
	}
	return nvi.KubeletVersion.Major() < clusterVersion.Major() ||
		nvi.KubeletVersion.Minor() < clusterVersion.Minor()
}

func (nvi NodeVersionInfo) ToleratesClusterVersion(clusterVersion *version.Version) bool {
	if nvi.IsControlPlane() {
		if nvi.APIServerVersion.Major() != clusterVersion.Major() ||
			nvi.APIServerVersion.Minor() != clusterVersion.Minor() {
			return false
		}
	}

	return nvi.KubeletVersion.Minor() == clusterVersion.Minor() ||
		nvi.KubeletVersion.Minor()+1 == clusterVersion.Minor()
}

// AllNodesVersioningInfo returns the version info for all nodes in the cluster
func AllNodesVersioningInfo() (NodeVersionInfoMap, error) {
	client, err := GetAdminClientSet()
	if err != nil {
		return NodeVersionInfoMap{}, errors.Wrap(err, "unable to get admin client set")
	}

	nodeList, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return NodeVersionInfoMap{}, errors.Wrap(err, "could not retrieve node list")
	}

	result := NodeVersionInfoMap{}

	for _, node := range nodeList.Items {
		nodeVersion, err := nodeVersioningInfoWithClientset(client, node.ObjectMeta.Name)
		if err != nil {
			return NodeVersionInfoMap{}, err
		}
		result[node.ObjectMeta.Name] = nodeVersion
	}

	return result, nil
}

// NodeVersioningInfo returns related versioning information about a node
func NodeVersioningInfo(nodeName string) (NodeVersionInfo, error) {
	client, err := GetAdminClientSet()
	if err != nil {
		return NodeVersionInfo{}, errors.Wrap(err, "unable to get admin client set")
	}

	nodeVersions, err := nodeVersioningInfoWithClientset(client, nodeName)
	if err != nil {
		return NodeVersionInfo{}, errors.Wrap(err, "unable to get node versioning info")
	}

	return nodeVersions, nil
}

func nodeVersioningInfoWithClientset(client kubernetes.Interface, nodeName string) (NodeVersionInfo, error) {
	nodeObject, err := client.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve node object")
	}

	kubeletVersion := version.MustParseSemantic(nodeObject.Status.NodeInfo.KubeletVersion)
	containerRuntimeVersionRaw := nodeObject.Status.NodeInfo.ContainerRuntimeVersion
	unschedulable := nodeObject.Spec.Unschedulable

	// Extract the container runtime version from the raw version
	containerRuntimeVersion := runtimeVersionRegexp.FindStringSubmatch(containerRuntimeVersionRaw)[1]

	nodeVersions := NodeVersionInfo{
		Nodename:                nodeName,
		ContainerRuntimeVersion: version.MustParseSemantic(containerRuntimeVersion),
		KubeletVersion:          kubeletVersion,
		Unschedulable:           unschedulable,
	}

	// find out the container image tags, depending on the role of the node
	if IsControlPlane(nodeObject) {
		var apiServerTag, controllerManagerTag, schedulerTag, etcdTag string
		apiServerTag, err = getPodContainerImageTagWithClientset(client, metav1.NamespaceSystem, fmt.Sprintf("%s-%s", "kube-apiserver", nodeName))
		if err != nil {
			return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve apiserver pod")
		}
		controllerManagerTag, err = getPodContainerImageTagWithClientset(client, metav1.NamespaceSystem, fmt.Sprintf("%s-%s", "kube-controller-manager", nodeName))
		if err != nil {
			return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve controller-manager pod")
		}
		schedulerTag, err = getPodContainerImageTagWithClientset(client, metav1.NamespaceSystem, fmt.Sprintf("%s-%s", "kube-scheduler", nodeName))
		if err != nil {
			return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve scheduler pod")
		}
		etcdTag, err = getPodContainerImageTagWithClientset(client, metav1.NamespaceSystem, fmt.Sprintf("%s-%s", "etcd", nodeName))
		if err != nil {
			return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve etcd pod")
		}

		nodeVersions.APIServerVersion = version.MustParseSemantic(apiServerTag)
		nodeVersions.ControllerManagerVersion = version.MustParseSemantic(controllerManagerTag)
		nodeVersions.SchedulerVersion = version.MustParseSemantic(schedulerTag)
		nodeVersions.EtcdVersion = version.MustParseSemantic(etcdTag)
	}

	return nodeVersions, nil
}

// AllWorkerNodesTolerateVersion checks that all schedulable worker nodes tolerate the given cluster version
func AllWorkerNodesTolerateVersion(clusterVersion *version.Version) (bool, error) {
	allNodesVersioningInfo, err := AllNodesVersioningInfo()
	if err != nil {
		return false, err
	}

	return allWorkerNodesTolerateVersionWithVersioningInfo(allNodesVersioningInfo, clusterVersion), nil
}

func allWorkerNodesTolerateVersionWithVersioningInfo(allNodesVersioningInfo NodeVersionInfoMap, clusterVersion *version.Version) bool {
	for _, nodeInfo := range allNodesVersioningInfo {
		if nodeInfo.IsControlPlane() {
			continue
		}
		if !nodeInfo.Unschedulable && !nodeInfo.ToleratesClusterVersion(clusterVersion) {
			return false
		}
	}
	return true
}

// AllControlPlanesMatchVersion checks that all control planes are on the same version, and that they match a cluster version
func AllControlPlanesMatchVersion(clusterVersion *version.Version) (bool, error) {
	allNodesVersioningInfo, err := AllNodesVersioningInfo()
	if err != nil {
		return false, err
	}
	return AllControlPlanesMatchVersionWithVersioningInfo(allNodesVersioningInfo, clusterVersion), nil
}

// AllControlPlanesMatchVersionWithVersioningInfo checks that all control planes are on the same version, and that they match a cluster version
// With the addition of passing in a NodeVersionInfoMap to make it easily testable
func AllControlPlanesMatchVersionWithVersioningInfo(allNodesVersioningInfo NodeVersionInfoMap, clusterVersion *version.Version) bool {
	for _, nodeInfo := range allNodesVersioningInfo {
		// skip workers
		if !nodeInfo.IsControlPlane() {
			continue
		}
		if !nodeInfo.ToleratesClusterVersion(clusterVersion) {
			return false
		}
	}
	return true
}

// AllNodesMatchVersionWithVersioningInfo returns if all nodes match a specific kubernetes version
// This can be used to determine e.g. if an upgrade is ongoing
func AllNodesMatchClusterVersionWithVersioningInfo(allNodesVersioningInfo NodeVersionInfoMap, clusterVersion *version.Version) bool {
	for _, version := range allNodesVersioningInfo {
		if !version.EqualsClusterVersion(clusterVersion) {
			return false
		}
	}
	return true
}
