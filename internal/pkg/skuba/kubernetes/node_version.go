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
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

var (
	// Container runtime is of the form `docker://version` or `cri-o://version`
	runtimeVersionRegexp = regexp.MustCompile(`//(.+)$`)
)

type NodeVersionInfo struct {
	Node                     *v1.Node
	ContainerRuntimeVersion  *version.Version
	KubeletVersion           *version.Version
	APIServerVersion         *version.Version
	ControllerManagerVersion *version.Version
	SchedulerVersion         *version.Version
	EtcdVersion              *version.Version
}

type NodeVersionInfoMap map[string]NodeVersionInfo

type VersionInquirer interface {
	AvailablePlatformVersions() []*version.Version
	NodeVersionInfoForClusterVersion(node *v1.Node, version *version.Version) NodeVersionInfo
}
type StaticVersionInquirer struct{}

func (si StaticVersionInquirer) AvailablePlatformVersions() []*version.Version {
	return AvailableVersionsForMap(supportedVersions)
}

func (si StaticVersionInquirer) NodeVersionInfoForClusterVersion(node *v1.Node, clusterVersion *version.Version) NodeVersionInfo {
	res := NodeVersionInfo{
		Node:                    node.DeepCopy(),
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

func (nvi NodeVersionInfo) Unschedulable() bool {
	return nvi.Node.Spec.Unschedulable
}

func (nvi NodeVersionInfo) IsControlPlane() bool {
	return IsControlPlane(nvi.Node)
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
func AllNodesVersioningInfo(client clientset.Interface) (NodeVersionInfoMap, error) {
	var lastErr error
	var nodeList *v1.NodeList
	err := wait.ExponentialBackoff(retry.DefaultBackoff, func() (bool, error) {
		var err error
		nodeList, err = client.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			lastErr = err
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return NodeVersionInfoMap{}, errors.Wrap(lastErr, "could not retrieve node list")
	}
	var podList *v1.PodList
	err = wait.ExponentialBackoff(retry.DefaultBackoff, func() (bool, error) {
		var err error
		podList, err = client.CoreV1().Pods(metav1.NamespaceSystem).List(metav1.ListOptions{})
		if err != nil {
			lastErr = err
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return NodeVersionInfoMap{}, errors.Wrap(lastErr, "could not retrieve pods")
	}

	result := NodeVersionInfoMap{}
	for _, node := range nodeList.Items {
		nodeVersion, err := nodeVersioningInfo(&node, podList)
		if err != nil {
			return NodeVersionInfoMap{}, err
		}
		result[node.ObjectMeta.Name] = nodeVersion
	}

	return result, nil
}

func nodeVersioningInfo(node *v1.Node, podList *v1.PodList) (NodeVersionInfo, error) {
	nodeName := node.ObjectMeta.Name
	kubeletVersion := version.MustParseSemantic(node.Status.NodeInfo.KubeletVersion)
	containerRuntimeVersionRaw := node.Status.NodeInfo.ContainerRuntimeVersion

	// Extract the container runtime version from the raw version
	containerRuntimeVersion, err := version.ParseSemantic(runtimeVersionRegexp.FindStringSubmatch(containerRuntimeVersionRaw)[1])
	if err != nil {
		return NodeVersionInfo{}, errors.Wrapf(err, "could not retrieve node %q container runtime version", nodeName)
	}

	nodeVersions := NodeVersionInfo{
		Node:                    node.DeepCopy(),
		ContainerRuntimeVersion: containerRuntimeVersion,
		KubeletVersion:          kubeletVersion,
	}

	// find out the container image tags, depending on the role of the node
	if IsControlPlane(node) {
		// check for empty pod list
		if len(podList.Items) == 0 {
			return NodeVersionInfo{}, errors.New("list of pods is empty")
		}
		// check that the needed pods exist
		apiserverPod, err := getPodFromPodList(podList, fmt.Sprintf("kube-apiserver-%s", nodeName))
		if err != nil {
			return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve api server pod")
		}
		controllerManagerPod, err := getPodFromPodList(podList, fmt.Sprintf("kube-controller-manager-%s", nodeName))
		if err != nil {
			return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve controller manager pod")
		}
		schedulerPod, err := getPodFromPodList(podList, fmt.Sprintf("kube-scheduler-%s", nodeName))
		if err != nil {
			return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve scheduler pod")
		}
		etcdPod, err := getPodFromPodList(podList, fmt.Sprintf("etcd-%s", nodeName))
		if err != nil {
			return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve etcd pod")
		}

		nodeVersions.APIServerVersion = version.MustParseSemantic(getPodContainerImageTagFromPodObject(apiserverPod))
		nodeVersions.ControllerManagerVersion = version.MustParseSemantic(getPodContainerImageTagFromPodObject(controllerManagerPod))
		nodeVersions.SchedulerVersion = version.MustParseSemantic(getPodContainerImageTagFromPodObject(schedulerPod))
		nodeVersions.EtcdVersion = version.MustParseSemantic(getPodContainerImageTagFromPodObject(etcdPod))
	}

	return nodeVersions, nil
}

// AllWorkerNodesTolerateVersion checks that all schedulable worker nodes tolerate the given cluster version
func AllWorkerNodesTolerateVersion(client clientset.Interface, clusterVersion *version.Version) (bool, error) {
	allNodesVersioningInfo, err := AllNodesVersioningInfo(client)
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
		if !nodeInfo.Unschedulable() && !nodeInfo.ToleratesClusterVersion(clusterVersion) {
			return false
		}
	}
	return true
}

// AllControlPlanesMatchVersion checks that all control planes are on the same version, and that they match a cluster version
func AllControlPlanesMatchVersion(client clientset.Interface, clusterVersion *version.Version) (bool, error) {
	allNodesVersioningInfo, err := AllNodesVersioningInfo(client)
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
