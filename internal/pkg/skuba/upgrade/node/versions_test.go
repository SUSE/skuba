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
	"fmt"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

type TestVersionInquirer struct {
	AvailableVersions kubernetes.KubernetesVersions
}

func (ti TestVersionInquirer) AvailablePlatformVersions() []*version.Version {
	return kubernetes.AvailableVersionsForMap(ti.AvailableVersions)
}

func (ti TestVersionInquirer) NodeVersionInfoForClusterVersion(node *v1.Node, clusterVersion *version.Version) kubernetes.NodeVersionInfo {
	res := kubernetes.NodeVersionInfo{
		Nodename:                node.ObjectMeta.Name,
		ContainerRuntimeVersion: version.MustParseSemantic(kubernetes.ComponentVersionWithAvailableVersions(kubernetes.ContainerRuntime, clusterVersion, ti.AvailableVersions)),
		KubeletVersion:          version.MustParseSemantic(kubernetes.ComponentVersionWithAvailableVersions(kubernetes.Kubelet, clusterVersion, ti.AvailableVersions)),
	}
	if kubernetes.IsControlPlane(node) {
		res.APIServerVersion = version.MustParseSemantic(kubernetes.ComponentVersionWithAvailableVersions(kubernetes.Hyperkube, clusterVersion, ti.AvailableVersions))
		res.ControllerManagerVersion = version.MustParseSemantic(kubernetes.ComponentVersionWithAvailableVersions(kubernetes.Hyperkube, clusterVersion, ti.AvailableVersions))
		res.SchedulerVersion = version.MustParseSemantic(kubernetes.ComponentVersionWithAvailableVersions(kubernetes.Hyperkube, clusterVersion, ti.AvailableVersions))
		res.EtcdVersion = version.MustParseSemantic(kubernetes.ComponentVersionWithAvailableVersions(kubernetes.Etcd, clusterVersion, ti.AvailableVersions))
	}
	return res
}

func isControlPlane(nodeVersionInfo kubernetes.NodeVersionInfo) bool {
	return nodeVersionInfo.IsControlPlane()
}

func TestNodesVersionAligned(t *testing.T) {
	var versions = []struct {
		name                   string
		currentClusterVersion  *version.Version
		allNodesVersioningInfo kubernetes.NodeVersionInfoMap
		nodeConsidered         func(kubernetes.NodeVersionInfo) bool
		expectedAligned        bool
	}{
		{
			name:                  "apiserver and kubelet aligned",
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			allNodesVersioningInfo: kubernetes.NodeVersionInfoMap{
				"cp1": {
					Nodename:         "cp1",
					APIServerVersion: version.MustParseSemantic("v1.14.0"),
					KubeletVersion:   version.MustParseSemantic("v1.14.0"),
				},
			},
			nodeConsidered:  isControlPlane,
			expectedAligned: true,
		},
		{
			name:                  "apiserver not aligned (by patch) and kubelet aligned",
			currentClusterVersion: version.MustParseSemantic("v1.14.9"),
			allNodesVersioningInfo: kubernetes.NodeVersionInfoMap{
				"cp1": {
					Nodename:         "cp1",
					APIServerVersion: version.MustParseSemantic("v1.14.0"),
					KubeletVersion:   version.MustParseSemantic("v1.14.9"),
				},
			},
			nodeConsidered:  isControlPlane,
			expectedAligned: true,
		},
		{
			name:                  "apiserver not aligned (by minor) and kubelet aligned",
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			allNodesVersioningInfo: kubernetes.NodeVersionInfoMap{
				"cp1": {
					Nodename:         "cp1",
					APIServerVersion: version.MustParseSemantic("v1.13.0"),
					KubeletVersion:   version.MustParseSemantic("v1.14.0"),
				},
			},
			nodeConsidered:  isControlPlane,
			expectedAligned: false,
		},
		{
			name:                  "apiserver aligned and kubelet not aligned (by patch)",
			currentClusterVersion: version.MustParseSemantic("v1.14.9"),
			allNodesVersioningInfo: kubernetes.NodeVersionInfoMap{
				"cp1": {
					Nodename:         "cp1",
					APIServerVersion: version.MustParseSemantic("v1.14.9"),
					KubeletVersion:   version.MustParseSemantic("v1.14.0"),
				},
			},
			nodeConsidered:  isControlPlane,
			expectedAligned: true,
		},
		{
			name:                  "apiserver aligned and kubelet not aligned (by minor)",
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			allNodesVersioningInfo: kubernetes.NodeVersionInfoMap{
				"cp1": {
					Nodename:         "cp1",
					APIServerVersion: version.MustParseSemantic("v1.14.0"),
					KubeletVersion:   version.MustParseSemantic("v1.13.0"),
				},
			},
			nodeConsidered:  isControlPlane,
			expectedAligned: false,
		},
		{
			name:                  "control plane aligned and worker node unaligned",
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			allNodesVersioningInfo: kubernetes.NodeVersionInfoMap{
				"cp1": {
					Nodename:         "cp1",
					APIServerVersion: version.MustParseSemantic("v1.14.0"),
					KubeletVersion:   version.MustParseSemantic("v1.14.0"),
				},
				"worker1": {
					Nodename:       "worker1",
					KubeletVersion: version.MustParseSemantic("v1.12.0"),
				},
			},
			nodeConsidered:  isControlPlane,
			expectedAligned: true,
		},
	}
	for _, tt := range versions {
		t.Run(tt.name, func(t *testing.T) {
			aligned := nodesVersionAligned(tt.currentClusterVersion, tt.allNodesVersioningInfo, tt.nodeConsidered)
			if aligned != tt.expectedAligned {
				t.Errorf("align result (%v) does not match expectation (%v)", aligned, tt.expectedAligned)
			}
		})
	}
}

type nodeRole uint

const (
	controlPlane nodeRole = iota
	worker
)

func nodeVersion(node string, nodeVersion string, nodeRole nodeRole) kubernetes.NodeVersionInfo {
	res := kubernetes.NodeVersionInfo{
		Nodename:                node,
		ContainerRuntimeVersion: version.MustParseSemantic(fmt.Sprintf("v%s", nodeVersion)),
		KubeletVersion:          version.MustParseSemantic(fmt.Sprintf("v%s", nodeVersion)),
	}
	if nodeRole == controlPlane {
		res.APIServerVersion = version.MustParseSemantic(fmt.Sprintf("v%s", nodeVersion))
		res.ControllerManagerVersion = version.MustParseSemantic(fmt.Sprintf("v%s", nodeVersion))
		res.SchedulerVersion = version.MustParseSemantic(fmt.Sprintf("v%s", nodeVersion))
		res.EtcdVersion = version.MustParseSemantic("3.3.11")
	}
	return res
}

func nodeVersionMap(controlPlaneNodes map[string]string, workerNodes map[string]string) kubernetes.NodeVersionInfoMap {
	res := kubernetes.NodeVersionInfoMap{}
	for node, version := range controlPlaneNodes {
		res[node] = nodeVersion(node, version, controlPlane)
	}
	for node, version := range workerNodes {
		res[node] = nodeVersion(node, version, worker)
	}
	return res
}

func controlPlaneNode(name string) v1.Node {
	return v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				kubeadmconstants.LabelNodeRoleMaster: "",
			},
		},
	}
}

func workerNode(name string) v1.Node {
	return v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func versionInquirer(versions ...string) kubernetes.VersionInquirer {
	res := TestVersionInquirer{
		AvailableVersions: kubernetes.KubernetesVersions{},
	}
	for _, version := range versions {
		res.AvailableVersions[version] = kubernetes.KubernetesVersion{
			ControlPlaneComponentsVersion: kubernetes.ControlPlaneComponentsVersion{
				HyperkubeVersion: fmt.Sprintf("v%s", version),
				EtcdVersion:      "3.3.11",
				CoreDNSVersion:   "1.2.6",
				PauseVersion:     "3.1",
			},
			ComponentsVersion: kubernetes.ComponentsVersion{
				ContainerRuntimeVersion: version,
				KubeletVersion:          version,
			},
			AddonsVersion: kubernetes.AddonsVersion{
				CiliumVersion:  "1.5.3",
				ToolingVersion: "0.1.0",
				KuredVersion:   "1.2.0",
			},
		}
	}
	return res
}

func TestControlPlaneUpdateStatusWithAvailableVersions(t *testing.T) {
	var versions = []struct {
		name                          string
		currentClusterVersion         *version.Version
		versionInquirer               kubernetes.VersionInquirer
		allNodesVersioningInfo        kubernetes.NodeVersionInfoMap
		node                          v1.Node
		expectedNodeVersionInfoUpdate NodeVersionInfoUpdate
		expectedError                 bool
	}{
		{
			name:                   "first control plane to be upgraded; no upgrades available",
			currentClusterVersion:  version.MustParseSemantic("v1.14.0"),
			versionInquirer:        versionInquirer("1.14.0"),
			allNodesVersioningInfo: nodeVersionMap(map[string]string{"cp1": "1.14.0"}, map[string]string{}),
			node:                   controlPlaneNode("cp1"),
			expectedNodeVersionInfoUpdate: NodeVersionInfoUpdate{
				Current: nodeVersion("cp1", "1.14.0", controlPlane),
				Update:  nodeVersion("cp1", "1.14.0", controlPlane),
			},
		},
		{
			name:                   "first control plane to be upgraded; upgrades available",
			currentClusterVersion:  version.MustParseSemantic("v1.14.0"),
			versionInquirer:        versionInquirer("1.14.0", "1.15.0"),
			allNodesVersioningInfo: nodeVersionMap(map[string]string{"cp1": "1.14.0"}, map[string]string{}),
			node:                   controlPlaneNode("cp1"),
			expectedNodeVersionInfoUpdate: NodeVersionInfoUpdate{
				Current: nodeVersion("cp1", "1.14.0", controlPlane),
				Update:  nodeVersion("cp1", "1.15.0", controlPlane),
			},
		},
		{
			name:                   "secondary control plane to be upgraded; upgrades available",
			currentClusterVersion:  version.MustParseSemantic("v1.15.0"),
			versionInquirer:        versionInquirer("1.14.0", "1.15.0"),
			allNodesVersioningInfo: nodeVersionMap(map[string]string{"cp1": "1.15.0", "cp2": "1.14.0"}, map[string]string{}),
			node:                   controlPlaneNode("cp2"),
			expectedNodeVersionInfoUpdate: NodeVersionInfoUpdate{
				Current: nodeVersion("cp2", "1.14.0", controlPlane),
				Update:  nodeVersion("cp2", "1.15.0", controlPlane),
			},
		},
		{
			name:                   "first control plane to be upgraded; outdated worker",
			currentClusterVersion:  version.MustParseSemantic("v1.14.0"),
			versionInquirer:        versionInquirer("1.13.0", "1.14.0", "1.15.0"),
			allNodesVersioningInfo: nodeVersionMap(map[string]string{"cp1": "1.14.0"}, map[string]string{"worker1": "1.13.0"}),
			node:                   controlPlaneNode("cp1"),
			expectedError:          true,
		},
	}
	for _, tt := range versions {
		t.Run(tt.name, func(t *testing.T) {
			nodeVersionInfoUpdate, err := controlPlaneUpdateStatusWithAvailableVersions(tt.currentClusterVersion, tt.allNodesVersioningInfo, &tt.node, tt.versionInquirer)
			if err == nil && tt.expectedError {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if err != nil && !tt.expectedError {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
			if !reflect.DeepEqual(nodeVersionInfoUpdate, tt.expectedNodeVersionInfoUpdate) {
				t.Errorf("returned version info update (%v) does not match the expected one (%v)", nodeVersionInfoUpdate, tt.expectedNodeVersionInfoUpdate)
			}
		})
	}
}

func TestWorkerUpdateStatusWithAvailableVersions(t *testing.T) {
	var versions = []struct {
		name                          string
		currentClusterVersion         *version.Version
		versionInquirer               kubernetes.VersionInquirer
		allNodesVersioningInfo        kubernetes.NodeVersionInfoMap
		node                          v1.Node
		expectedNodeVersionInfoUpdate NodeVersionInfoUpdate
		expectedError                 bool
	}{
		{
			name:                   "worker; no upgrades available",
			currentClusterVersion:  version.MustParseSemantic("v1.14.0"),
			versionInquirer:        versionInquirer("1.14.0"),
			allNodesVersioningInfo: nodeVersionMap(map[string]string{}, map[string]string{"worker1": "1.14.0"}),
			node:                   workerNode("worker1"),
			expectedNodeVersionInfoUpdate: NodeVersionInfoUpdate{
				Current: nodeVersion("worker1", "1.14.0", worker),
				Update:  nodeVersion("worker1", "1.14.0", worker),
			},
		},
		{
			name:                   "worker with outdated control plane; upgrades available",
			currentClusterVersion:  version.MustParseSemantic("v1.15.0"),
			versionInquirer:        versionInquirer("1.14.0", "1.15.0"),
			allNodesVersioningInfo: nodeVersionMap(map[string]string{"cp1": "1.15.0", "cp2": "1.14.0"}, map[string]string{"worker1": "1.14.0"}),
			node:                   workerNode("worker1"),
			expectedError:          true,
		},
	}
	for _, tt := range versions {
		t.Run(tt.name, func(t *testing.T) {
			nodeVersionInfoUpdate, err := workerUpdateStatusWithAvailableVersions(tt.currentClusterVersion, tt.allNodesVersioningInfo, &tt.node, tt.versionInquirer)
			if err == nil && tt.expectedError {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if err != nil && !tt.expectedError {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
			if !reflect.DeepEqual(nodeVersionInfoUpdate, tt.expectedNodeVersionInfoUpdate) {
				t.Errorf("returned version info update (%v) does not match the expected one (%v)", nodeVersionInfoUpdate, tt.expectedNodeVersionInfoUpdate)
			}
		})
	}
}
