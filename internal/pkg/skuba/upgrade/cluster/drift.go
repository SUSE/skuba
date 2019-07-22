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

package cluster

import (
	"k8s.io/apimachinery/pkg/util/version"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

func driftedNodesWithVersions(currentClusterVersion *version.Version, nodesVersionInfo kubernetes.NodeVersionInfoMap) []kubernetes.NodeVersionInfo {
	driftedNodes := []kubernetes.NodeVersionInfo{}
	for _, node := range nodesVersionInfo {
		if node.Unschedulable {
			continue
		}
		if node.DriftsFromClusterVersion(currentClusterVersion) {
			driftedNodes = append(driftedNodes, node)
		}
	}
	return driftedNodes
}

// DriftedNodes return the list of outdated nodes with regards to the current cluster
// version. Unschedulable nodes will be ignored. Only nodes that have a lower minor or
// major version than the current cluster version are considered. If the difference on
// the node version with regards to the current cluster version is only the patch level
// version, the node won't be included in the list.
func DriftedNodes() ([]kubernetes.NodeVersionInfo, error) {
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion()
	if err != nil {
		return []kubernetes.NodeVersionInfo{}, err
	}
	allNodesVersioningInfo, err := kubernetes.AllNodesVersioningInfo()
	if err != nil {
		return []kubernetes.NodeVersionInfo{}, err
	}

	return driftedNodesWithVersions(currentClusterVersion, allNodesVersioningInfo), nil
}
