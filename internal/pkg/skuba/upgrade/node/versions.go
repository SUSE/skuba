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

package node

import (
	"reflect"

	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	upgradecluster "github.com/SUSE/skuba/internal/pkg/skuba/upgrade/cluster"
)

type NodeVersionInfoUpdate struct {
	Current kubernetes.NodeVersionInfo
	Update  kubernetes.NodeVersionInfo
}

func (nviu NodeVersionInfoUpdate) IsUpdated() bool {
	return reflect.DeepEqual(nviu.Current, nviu.Update)
}

func (nviu NodeVersionInfoUpdate) IsFirstControlPlaneNodeToBeUpgraded() bool {
	clientSet, _ := kubernetes.GetAdminClientSet()
	isControlPlane := nviu.Current.IsControlPlane()
	currentClusterVersion, _ := kubeadm.GetCurrentClusterVersion(clientSet)
	allControlPlanesMatchVersion, _ := kubernetes.AllControlPlanesMatchVersion(currentClusterVersion)
	matchesClusterVersion := (currentClusterVersion.String() == nviu.Current.KubeletVersion.String())

	return isControlPlane && allControlPlanesMatchVersion && matchesClusterVersion
}

func UpdateStatus(nodeName string) (NodeVersionInfoUpdate, error) {
	clientSet, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return NodeVersionInfoUpdate{}, errors.Wrap(err, "unable to get admin client set")
	}
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(clientSet)
	if err != nil {
		return NodeVersionInfoUpdate{}, err
	}
	allNodesVersioningInfo, err := kubernetes.AllNodesVersioningInfo()
	if err != nil {
		return NodeVersionInfoUpdate{}, err
	}
	node, err := clientSet.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return NodeVersionInfoUpdate{}, errors.Wrapf(err, "could not find node %s", nodeName)
	}
	if kubernetes.IsControlPlane(node) {
		return controlPlaneUpdateStatus(currentClusterVersion, allNodesVersioningInfo, node)
	}
	return workerUpdateStatus(currentClusterVersion, allNodesVersioningInfo, node)
}

func nodesVersionAligned(version *version.Version, allNodesVersioningInfo kubernetes.NodeVersionInfoMap, nodeConsidered func(kubernetes.NodeVersionInfo) bool) bool {
	for _, nodeInfo := range allNodesVersioningInfo {
		if nodeConsidered(nodeInfo) && nodeInfo.DriftsFromClusterVersion(version) {
			return false
		}
	}
	return true
}

func isSchedulableWorkerNode(nodeVersionInfo kubernetes.NodeVersionInfo) bool {
	return !nodeVersionInfo.Unschedulable && !nodeVersionInfo.IsControlPlane()
}

func controlPlaneUpdateStatus(currentClusterVersion *version.Version, allNodesVersioningInfo kubernetes.NodeVersionInfoMap, node *v1.Node) (NodeVersionInfoUpdate, error) {
	return controlPlaneUpdateStatusWithAvailableVersions(currentClusterVersion, allNodesVersioningInfo, node, kubernetes.StaticVersionInquirer{})
}

func controlPlaneUpdateStatusWithAvailableVersions(currentClusterVersion *version.Version, allNodesVersioningInfo kubernetes.NodeVersionInfoMap, node *v1.Node, versionInquirer kubernetes.VersionInquirer) (NodeVersionInfoUpdate, error) {
	// There are two different cases for control plane upgrade:
	//   1. This is the first control plane to be upgraded
	//     1.1. All control planes and schedulable worker nodes are in the same version
	//     1.2. There's a new platform version available
	//   2. This is a secondary control plane to be upgraded
	//     2.1. The current cluster version is newer than the control plane component versions in this node
	//     2.2. All schedulable worker nodes are at this control plane version
	nodeVersioningInfo, ok := allNodesVersioningInfo[node.ObjectMeta.Name]
	if !ok {
		return NodeVersionInfoUpdate{}, errors.New("could not find node on the list of all nodes")
	}
	if nodeVersioningInfo.LessThanClusterVersion(currentClusterVersion) {
		// Second case, the current cluster version was bumped by another control plane that got upgraded
		// first
		return NodeVersionInfoUpdate{
			Current: nodeVersioningInfo,
			Update:  versionInquirer.NodeVersionInfoForClusterVersion(node, currentClusterVersion),
		}, nil
	}
	// Either there are no platform updates available, or we are in the first case (upgrading the first
	// control plane)
	upgradePath, err := upgradecluster.UpgradePathWithAvailableVersions(currentClusterVersion, versionInquirer.AvailablePlatformVersions())
	if err != nil {
		return NodeVersionInfoUpdate{}, errors.New("could not determine if a new version of the platform is available")
	}
	if len(upgradePath) > 0 {
		// There are platform updates available, return the next version bump if all schedulable
		// worker nodes are not already drifting
		if !nodesVersionAligned(currentClusterVersion, allNodesVersioningInfo, isSchedulableWorkerNode) {
			return NodeVersionInfoUpdate{}, errors.New("at least one schedulable worker node has drifted behind, upgrading this node would imply that they wouldn't be able to communicate with the updated version of this node. Please, upgrade that node, cordon it or remove it from the cluster")
		}
		return NodeVersionInfoUpdate{
			Current: nodeVersioningInfo,
			Update:  versionInquirer.NodeVersionInfoForClusterVersion(node, upgradePath[0]),
		}, nil
	}
	if !nodeVersioningInfo.EqualsClusterVersion(currentClusterVersion) {
		return NodeVersionInfoUpdate{}, errors.Errorf("cannot infer how to upgrade this node from version %s to version %s", nodeVersioningInfo.String(), currentClusterVersion.String())
	}
	// This node is up to date and there are not newer versions available of the platform
	return NodeVersionInfoUpdate{
		Current: nodeVersioningInfo,
		Update:  nodeVersioningInfo,
	}, nil
}

func workerUpdateStatus(currentClusterVersion *version.Version, allNodesVersioningInfo kubernetes.NodeVersionInfoMap, node *v1.Node) (NodeVersionInfoUpdate, error) {
	return workerUpdateStatusWithAvailableVersions(currentClusterVersion, allNodesVersioningInfo, node, kubernetes.StaticVersionInquirer{})
}

func workerUpdateStatusWithAvailableVersions(currentClusterVersion *version.Version, allNodesVersioningInfo kubernetes.NodeVersionInfoMap, node *v1.Node, versionInquirer kubernetes.VersionInquirer) (NodeVersionInfoUpdate, error) {
	// Check that all control plane nodes have at least the current cluster version we plan to
	// upgrade this worker node to. If not, they need to be fully upgraded first
	controlPlanesMatchVersion := kubernetes.AllControlPlanesMatchVersionWithVersioningInfo(allNodesVersioningInfo, currentClusterVersion)
	if !controlPlanesMatchVersion {
		return NodeVersionInfoUpdate{}, errors.New("at least one control plane does not tolerate the current cluster version")
	}

	// Worker nodes only update themselves to the `currentClusterVersion`. They get updated after
	// control planes, that bump the current cluster version when the first control plane is updated
	var ok bool
	res := NodeVersionInfoUpdate{}
	res.Current, ok = allNodesVersioningInfo[node.ObjectMeta.Name]
	if !ok {
		return NodeVersionInfoUpdate{}, errors.New("could not find node on the list of all nodes")
	}
	res.Update = versionInquirer.NodeVersionInfoForClusterVersion(node, currentClusterVersion)
	return res, nil
}
