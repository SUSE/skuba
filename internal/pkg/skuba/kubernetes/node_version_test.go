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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes/fake"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

const (
	master0           = "master-0"
	master1           = "master-1"
	worker0           = "worker-0"
	worker1           = "worker-1"
	apiserver         = "kube-apiserver"
	controllerManager = "kube-controller-manager"
	scheduler         = "kube-scheduler"
	etcd              = "etcd"
)

func TestAvailablePlatformVersions(t *testing.T) {
	versions := StaticVersionInquirer{}.AvailablePlatformVersions()
	for _, v := range versions {
		t.Run(fmt.Sprintf("version(%v) should parse semantic", v), func(t *testing.T) {
			if _, err := version.ParseSemantic(v.String()); err != nil {
				t.Errorf("error not expected: (%v)", err)
			}
		})
	}
}

func TestNodeVersionInfoForClusterVersion(t *testing.T) {
	tests := []struct {
		node corev1.Node
	}{
		{
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: master0,
					Labels: map[string]string{
						kubeadmconstants.LabelNodeRoleMaster: "",
					},
				},
			},
		},
		{
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: worker0,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		var ver string
		isMaster := IsControlPlane(&tt.node)
		clusterVersions := StaticVersionInquirer{}.AvailablePlatformVersions()
		for _, cv := range clusterVersions {
			tName := fmt.Sprintf("%v node version info when cluster version is %v", tt.node.Name, cv)
			t.Run(tName, func(t *testing.T) {
				vInfo := StaticVersionInquirer{}.NodeVersionInfoForClusterVersion(&tt.node, cv)
				if vInfo.Nodename == "" {
					t.Error("node name expected, but none returned")
				}
				ver = vInfo.ContainerRuntimeVersion.String()
				if _, err := version.ParseSemantic(ver); err != nil {
					t.Errorf("container runtime version(%v) should parse semantic", ver)
				}
				ver = vInfo.KubeletVersion.String()
				if _, err := version.ParseSemantic(ver); err != nil {
					t.Errorf("kubelet version(%v) should parse semantic", ver)
				}

				if isMaster {
					ver = vInfo.APIServerVersion.String()
					if _, err := version.ParseSemantic(ver); err != nil {
						t.Errorf("api server version(%v) should parse semantic", ver)
					}
					ver = vInfo.ControllerManagerVersion.String()
					if _, err := version.ParseSemantic(ver); err != nil {
						t.Errorf("controller manager version(%v) should parse semantic", ver)
					}
					ver = vInfo.SchedulerVersion.String()
					if _, err := version.ParseSemantic(ver); err != nil {
						t.Errorf("scheduler version(%v) should parse semantic", ver)
					}
					ver = vInfo.EtcdVersion.String()
					if _, err := version.ParseSemantic(ver); err != nil {
						t.Errorf("etcd version(%v) should parse semantic", ver)
					}
				}
			})
		}
	}
}

func TestString(t *testing.T) {
	cluster := "1.1.1"
	clusterVersion := version.MustParseSemantic(cluster)
	tests := []struct {
		name            string
		nodeVersionInfo *NodeVersionInfo
	}{
		{
			name: "master version",
			nodeVersionInfo: &NodeVersionInfo{
				APIServerVersion: clusterVersion,
			},
		},
		{
			name: "woker version",
			nodeVersionInfo: &NodeVersionInfo{
				KubeletVersion: clusterVersion,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actualReturn := tt.nodeVersionInfo.String()
			actualType := reflect.TypeOf(actualReturn).Kind()
			if actualType != reflect.String {
				t.Errorf("expect type string: %v", actualType)
			}
			if actualReturn != cluster {
				t.Errorf("got: (%v), want: (%v)", actualReturn, cluster)
			}
		})
	}
}

func TestEqualsClusterVersion(t *testing.T) {
	clusterVersion := version.MustParseSemantic("v1.1.1")
	wrongVersion := version.MustParseSemantic("v1.1.0")
	tests := []struct {
		name             string
		isMaster         bool
		apiServerVersion *version.Version
		kubeletVersion   *version.Version
		clusterVersion   *version.Version
		expectReturn     bool
	}{
		{
			name:             "master node (apiserver, kubelet) version equal to cluster version",
			isMaster:         true,
			apiServerVersion: clusterVersion,
			kubeletVersion:   clusterVersion,
			clusterVersion:   clusterVersion,
			expectReturn:     true,
		},
		{
			name:             "master node (kubelet) version not equal to cluster version",
			isMaster:         true,
			apiServerVersion: clusterVersion,
			kubeletVersion:   wrongVersion,
			clusterVersion:   clusterVersion,
			expectReturn:     false,
		},
		{
			name:             "master node (apiserver, kubelet) version not equal to cluster version",
			isMaster:         true,
			apiServerVersion: wrongVersion,
			kubeletVersion:   wrongVersion,
			clusterVersion:   clusterVersion,
			expectReturn:     false,
		},
		{
			name:           "worker node (kubelet) version equal to cluster version",
			kubeletVersion: clusterVersion,
			clusterVersion: clusterVersion,
			expectReturn:   true,
		},
		{
			name:           "worker node (kubelet) version not equal to cluster version",
			kubeletVersion: wrongVersion,
			clusterVersion: clusterVersion,
			expectReturn:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var nodeVersionInfo NodeVersionInfo
			switch tt.isMaster {
			case true:
				nodeVersionInfo = NodeVersionInfo{
					APIServerVersion: tt.apiServerVersion,
					KubeletVersion:   tt.kubeletVersion,
				}
			default:
				nodeVersionInfo = NodeVersionInfo{
					KubeletVersion: tt.kubeletVersion,
				}
			}
			actualReturn := nodeVersionInfo.EqualsClusterVersion(tt.clusterVersion)
			if actualReturn != tt.expectReturn {
				t.Errorf("got: (%v), want: (%v)", actualReturn, tt.expectReturn)
			}
		})
	}
}

func TestLessThanClusterVersion(t *testing.T) {
	clusterVersionPlus := version.MustParseSemantic("v1.1.2")
	clusterVersion := version.MustParseSemantic("v1.1.1")
	clusterVersionMinus := version.MustParseSemantic("v1.1.0")
	tests := []struct {
		name             string
		isMaster         bool
		apiServerVersion *version.Version
		kubeletVersion   *version.Version
		clusterVersion   *version.Version
		expectReturn     bool
	}{
		{
			name:             "master node version less than cluster version",
			isMaster:         true,
			apiServerVersion: clusterVersionMinus,
			kubeletVersion:   clusterVersionMinus,
			clusterVersion:   clusterVersion,
			expectReturn:     true,
		},
		{
			name:             "master node (kubelet) version less than cluster version",
			isMaster:         true,
			apiServerVersion: clusterVersion,
			kubeletVersion:   clusterVersionMinus,
			clusterVersion:   clusterVersion,
			expectReturn:     true,
		},
		{
			name:             "master node version more than cluster version",
			isMaster:         true,
			apiServerVersion: clusterVersionPlus,
			kubeletVersion:   clusterVersionPlus,
			clusterVersion:   clusterVersion,
			expectReturn:     false,
		},
		{
			name:           "worker node version less than cluster version",
			kubeletVersion: clusterVersionMinus,
			clusterVersion: clusterVersion,
			expectReturn:   true,
		},
		{
			name:           "worker node version more than cluster version",
			kubeletVersion: clusterVersionPlus,
			clusterVersion: clusterVersion,
			expectReturn:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var nodeVersionInfo NodeVersionInfo
			switch tt.isMaster {
			case true:
				nodeVersionInfo = NodeVersionInfo{
					APIServerVersion: tt.apiServerVersion,
					KubeletVersion:   tt.kubeletVersion,
				}
			default:
				nodeVersionInfo = NodeVersionInfo{
					KubeletVersion: tt.kubeletVersion,
				}
			}
			actualReturn := nodeVersionInfo.LessThanClusterVersion(tt.clusterVersion)
			if actualReturn != tt.expectReturn {
				t.Errorf("got: (%v), want: (%v)", actualReturn, tt.expectReturn)
			}
		})
	}
}

func TestDriftsFromClusterVersion(t *testing.T) {
	clusterVersion := version.MustParseSemantic("v1.1.1")
	majorLess := version.MustParseSemantic("v0.1.1")
	minorLess := version.MustParseSemantic("v1.0.1")
	patchLess := version.MustParseSemantic("v1.1.0")
	tests := []struct {
		name             string
		isMaster         bool
		clusterVersion   *version.Version
		apiServerVersion *version.Version
		kubeletVersion   *version.Version
		expectReturn     bool
	}{
		{
			name:             "master node version same as cluster version",
			isMaster:         true,
			clusterVersion:   clusterVersion,
			apiServerVersion: clusterVersion,
			kubeletVersion:   clusterVersion,
			expectReturn:     false,
		},
		{
			name:             "master node (apiserver) major version less than cluster version",
			isMaster:         true,
			clusterVersion:   clusterVersion,
			apiServerVersion: majorLess,
			expectReturn:     true,
		},
		{
			name:             "master node (apiserver) minor version less than cluster version",
			isMaster:         true,
			apiServerVersion: minorLess,
			clusterVersion:   clusterVersion,
			expectReturn:     true,
		},
		{
			name:             "master node (kubelet) major version less than cluster version",
			isMaster:         true,
			clusterVersion:   clusterVersion,
			apiServerVersion: clusterVersion,
			kubeletVersion:   majorLess,
			expectReturn:     true,
		},
		{
			name:             "master node (kubelet) minor version less than cluster version",
			isMaster:         true,
			clusterVersion:   clusterVersion,
			apiServerVersion: clusterVersion,
			kubeletVersion:   minorLess,
			expectReturn:     true,
		},
		{
			name:             "master node patch version less than cluster version",
			isMaster:         true,
			clusterVersion:   clusterVersion,
			apiServerVersion: patchLess,
			kubeletVersion:   patchLess,
			expectReturn:     false,
		},
		{
			name:           "worker node same as cluster version",
			clusterVersion: clusterVersion,
			kubeletVersion: clusterVersion,
			expectReturn:   false,
		},
		{
			name:           "worker node major version less than cluster version",
			clusterVersion: clusterVersion,
			kubeletVersion: majorLess,
			expectReturn:   true,
		},
		{
			name:           "worker node minor version less than cluster version",
			clusterVersion: clusterVersion,
			kubeletVersion: minorLess,
			expectReturn:   true,
		},
		{
			name:           "worker node patch version less than cluster version",
			clusterVersion: clusterVersion,
			kubeletVersion: patchLess,
			expectReturn:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var nodeVersionInfo NodeVersionInfo
			switch tt.isMaster {
			case true:
				nodeVersionInfo = NodeVersionInfo{
					APIServerVersion: tt.apiServerVersion,
					KubeletVersion:   tt.kubeletVersion,
				}
			default:
				nodeVersionInfo = NodeVersionInfo{
					KubeletVersion: tt.kubeletVersion,
				}
			}
			actualReturn := nodeVersionInfo.DriftsFromClusterVersion(tt.clusterVersion)
			if actualReturn != tt.expectReturn {
				t.Errorf("got: (%v), want: (%v)", actualReturn, tt.expectReturn)
			}
		})
	}
}

func TestToleratesFromClusterVersion(t *testing.T) {
	clusterVersion := version.MustParseSemantic("v1.1.1")
	majorLess := version.MustParseSemantic("v0.1.1")
	majorMore := version.MustParseSemantic("v2.1.1")
	minorLess := version.MustParseSemantic("v1.0.1")
	minorMore := version.MustParseSemantic("v1.2.1")
	patchLess := version.MustParseSemantic("v1.1.0")
	patchMore := version.MustParseSemantic("v1.1.2")
	tests := []struct {
		name             string
		isMaster         bool
		apiServerVersion *version.Version
		kubeletVersion   *version.Version
		expectReturn     bool
	}{
		{
			name:             "master node version same as cluster version",
			isMaster:         true,
			apiServerVersion: clusterVersion,
			kubeletVersion:   clusterVersion,
			expectReturn:     true,
		},
		{
			name:             "master node (apiserver) minor version is one less than cluster version",
			isMaster:         true,
			apiServerVersion: minorLess,
			expectReturn:     false,
		},
		{
			name:             "master node (apiserver) minor version is one greater than cluster version",
			isMaster:         true,
			apiServerVersion: minorMore,
			expectReturn:     false,
		},
		{
			name:             "master node (apiserver) major version is one less than cluster version",
			isMaster:         true,
			apiServerVersion: majorLess,
			expectReturn:     false,
		},
		{
			name:             "master node (apiserver) major version is one greater than cluster version",
			isMaster:         true,
			apiServerVersion: majorMore,
			expectReturn:     false,
		},
		{
			name:             "master node (kubelet) minor version is one less than cluster version",
			isMaster:         true,
			apiServerVersion: clusterVersion,
			kubeletVersion:   minorLess,
			expectReturn:     true,
		},
		{
			name:             "master node (kubelet) minor version is one greater than cluster version",
			isMaster:         true,
			apiServerVersion: clusterVersion,
			kubeletVersion:   minorMore,
			expectReturn:     false,
		},
		{
			name:             "master node patch version is one less than as cluster version",
			isMaster:         true,
			apiServerVersion: patchLess,
			kubeletVersion:   patchLess,
			expectReturn:     true,
		},
		{
			name:             "master node patch version is one greater than as cluster version",
			isMaster:         true,
			apiServerVersion: patchMore,
			kubeletVersion:   patchMore,
			expectReturn:     true,
		},
		{
			name:           "worker node version is same as cluster version",
			kubeletVersion: clusterVersion,
			expectReturn:   true,
		},
		{
			name:           "worker node minor version is one less than cluster version",
			kubeletVersion: minorLess,
			expectReturn:   true,
		},
		{
			name:           "worker node minor version is one greater than cluster version",
			kubeletVersion: minorMore,
			expectReturn:   false,
		},
		{
			name:           "worker node patch version is one less than cluster version",
			kubeletVersion: patchLess,
			expectReturn:   true,
		},
		{
			name:           "worker node patch version is one greater than cluster version",
			kubeletVersion: patchMore,
			expectReturn:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var nodeVersionInfo NodeVersionInfo
			switch tt.isMaster {
			case true:
				nodeVersionInfo = NodeVersionInfo{
					APIServerVersion: tt.apiServerVersion,
					KubeletVersion:   tt.kubeletVersion,
				}
			default:
				nodeVersionInfo = NodeVersionInfo{
					KubeletVersion: tt.kubeletVersion,
				}
			}
			actualReturn := nodeVersionInfo.ToleratesClusterVersion(clusterVersion)
			if actualReturn != tt.expectReturn {
				t.Errorf("got: (%v), want: (%v)", actualReturn, tt.expectReturn)
			}
		})
	}
}

func createFakeMasterNode(nodeName string, kubeletVersion string, crioVersion string, unschedulable bool) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Labels: map[string]string{
				kubeadmconstants.LabelNodeRoleMaster: "",
			},
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion:          kubeletVersion,
				ContainerRuntimeVersion: "cri-o://" + crioVersion,
			},
		},
		Spec: corev1.NodeSpec{
			Unschedulable: unschedulable,
		},
	}
}

func createFakeWorkerNode(nodeName string, kubeletVersion string, unschedulable bool) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				KubeletVersion:          kubeletVersion,
				ContainerRuntimeVersion: "cri-o://1.1.1",
			},
		},
		Spec: corev1.NodeSpec{
			Unschedulable: unschedulable,
		},
	}
}

func createFakePod(podName string, nodeName string, version string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName + "-" + nodeName,
			Namespace: metav1.NamespaceSystem,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  podName,
					Image: "registry.suse.com/caasp/v4/" + podName + ":" + version,
				},
			},
			NodeName: nodeName,
		},
	}
}

func TestNodeVersioningInfo(t *testing.T) {
	cluster := "1.14.1"
	clusterVersion := version.MustParseSemantic(cluster)
	testVersion := version.MustParseSemantic("2.2.2")
	node := master0
	var tests = []struct {
		name                     string
		unschedulable            bool
		kubeletVersion           *version.Version
		apiServerVersion         *version.Version
		controllerManagerVersion *version.Version
		schedulerVersion         *version.Version
		etcdVersion              *version.Version
		containerRuntimeVersion  string
		apiServerNode            string
		controllerManagerNode    string
		schedulerNode            string
		etcdNode                 string
		expectError              bool
	}{
		{
			name:                     "node version info schedulable",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            node,
			controllerManagerNode:    node,
			schedulerNode:            node,
			etcdNode:                 node,
			unschedulable:            false,
		},
		{
			name:                     "node version info unschedulable",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            node,
			controllerManagerNode:    node,
			schedulerNode:            node,
			etcdNode:                 node,
			unschedulable:            true,
		},
		{
			name:                     "node version info apiserver",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         testVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            node,
			controllerManagerNode:    node,
			schedulerNode:            node,
			etcdNode:                 node,
			unschedulable:            false,
		},
		{
			name:                     "node version info controller manager",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: testVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            node,
			controllerManagerNode:    node,
			schedulerNode:            node,
			etcdNode:                 node,
			unschedulable:            false,
		},
		{
			name:                     "node version info scheduler",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         testVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            node,
			controllerManagerNode:    node,
			schedulerNode:            node,
			etcdNode:                 node,
			unschedulable:            false,
		},
		{
			name:                     "node version info etcd",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              testVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            node,
			controllerManagerNode:    node,
			schedulerNode:            node,
			etcdNode:                 node,
			unschedulable:            false,
		},
		{
			name:                     "container runtime version unknown",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  "Unknown",
			apiServerNode:            node,
			controllerManagerNode:    node,
			schedulerNode:            node,
			etcdNode:                 node,
			expectError:              true,
		},
		{
			name:                     "missing apiserver",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            "missing",
			controllerManagerNode:    node,
			schedulerNode:            node,
			etcdNode:                 node,
			expectError:              true,
		},
		{
			name:                     "missing control manager",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            node,
			controllerManagerNode:    "missing",
			schedulerNode:            node,
			etcdNode:                 node,
			expectError:              true,
		},
		{
			name:                     "missing scheduler",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            node,
			controllerManagerNode:    node,
			schedulerNode:            "missing",
			etcdNode:                 node,
			expectError:              true,
		},
		{
			name:                     "missing etcd",
			kubeletVersion:           clusterVersion,
			apiServerVersion:         clusterVersion,
			controllerManagerVersion: clusterVersion,
			schedulerVersion:         clusterVersion,
			etcdVersion:              clusterVersion,
			containerRuntimeVersion:  cluster,
			apiServerNode:            node,
			controllerManagerNode:    node,
			schedulerNode:            node,
			etcdNode:                 "missing",
			expectError:              true,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						createFakeMasterNode(node, tt.kubeletVersion.String(), tt.containerRuntimeVersion, tt.unschedulable),
					},
				},
				&corev1.PodList{
					Items: []corev1.Pod{
						createFakePod(apiserver, tt.apiServerNode, tt.apiServerVersion.String()),
						createFakePod(controllerManager, tt.controllerManagerNode, tt.controllerManagerVersion.String()),
						createFakePod(scheduler, tt.schedulerNode, tt.schedulerVersion.String()),
						createFakePod(etcd, tt.etcdNode, tt.etcdVersion.String()),
					},
				},
			)

			actualReturn, err := NodeVersioningInfo(clientset, node)

			switch tt.expectError {
			case true:
				if err == nil {
					t.Error("error expected, but no error reported")
				}
			default:
				crv, err := version.ParseSemantic(tt.containerRuntimeVersion)
				if err != nil {
					t.Errorf("error not expected: (%v)", err)
				}
				expectReturn := NodeVersionInfo{
					Nodename:                 master0,
					ContainerRuntimeVersion:  crv,
					KubeletVersion:           tt.kubeletVersion,
					APIServerVersion:         tt.apiServerVersion,
					ControllerManagerVersion: tt.controllerManagerVersion,
					SchedulerVersion:         tt.schedulerVersion,
					EtcdVersion:              tt.etcdVersion,
					Unschedulable:            tt.unschedulable,
				}
				if !reflect.DeepEqual(actualReturn, expectReturn) {
					t.Errorf("got: (%v), want: (%v)", actualReturn, expectReturn)
				}
			}
		})
	}
}

func TestAllWorkerNodesTolerateVersion(t *testing.T) {
	clusterVersion := "1.14.1"
	cluster := version.MustParseSemantic(clusterVersion)
	tolerateVersion := "1.13.1"
	wrongVersion := "1.12.1"
	var nodes = []struct {
		name          string
		fakeClientset *fake.Clientset
		expectReturn  bool
		expectError   bool
	}{
		{
			name: "should not check masters",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						createFakeMasterNode(master0, wrongVersion, clusterVersion, false),
					},
				},
				&corev1.PodList{
					Items: []corev1.Pod{
						createFakePod(apiserver, master0, wrongVersion),
						createFakePod(controllerManager, master0, clusterVersion),
						createFakePod(scheduler, master0, clusterVersion),
						createFakePod(etcd, master0, clusterVersion),
					},
				},
			),
			expectReturn: true,
		},
		{
			name: "all workers equals cluster version",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						createFakeWorkerNode(worker0, clusterVersion, false),
						createFakeWorkerNode(worker1, clusterVersion, false),
					},
				},
			),
			expectReturn: true,
		},
		{
			name: "all workers tolerates cluster version",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						createFakeWorkerNode(worker0, tolerateVersion, false),
						createFakeWorkerNode(worker1, tolerateVersion, false),
					},
				},
			),
			expectReturn: true,
		},
		{
			name: "a worker needs update",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						createFakeWorkerNode(worker0, clusterVersion, false),
						createFakeWorkerNode(worker1, wrongVersion, false),
					},
				},
			),
			expectReturn: false,
		},
		{
			name: "all workers tolerate except an unschedulable one",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						createFakeWorkerNode(worker0, tolerateVersion, false),
						createFakeWorkerNode(worker1, tolerateVersion, true),
					},
				},
			),
			expectReturn: true,
		},
	}
	for _, tt := range nodes {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actualReturn, err := AllWorkerNodesTolerateVersion(tt.fakeClientset, cluster)
			if err != nil {
				t.Errorf("error not expected: (%v)", err)
				return
			}
			if actualReturn != tt.expectReturn {
				t.Errorf("got: (%v), want: (%v)", actualReturn, tt.expectReturn)
			}
		})
	}
}

func TestAllControlPlanesMatchVersionWithVersioningInfo(t *testing.T) {
	clusterVersion := "1.14.1"
	cluster := version.MustParseSemantic(clusterVersion)
	wrongVersion := "1.13.1"
	var nodes = []struct {
		name          string
		fakeClientset *fake.Clientset
		expectReturn  bool
	}{
		{
			name: "should not check workers",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						createFakeWorkerNode(worker0, wrongVersion, false),
					},
				},
			),
			expectReturn: true,
		},
		{
			name: "all masters match cluster version",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						createFakeMasterNode(master0, clusterVersion, clusterVersion, false),
						createFakeMasterNode(master1, clusterVersion, clusterVersion, false),
					},
				},
				&corev1.PodList{
					Items: []corev1.Pod{
						// master0
						createFakePod(apiserver, master0, clusterVersion),
						createFakePod(controllerManager, master0, clusterVersion),
						createFakePod(scheduler, master0, clusterVersion),
						createFakePod(etcd, master0, clusterVersion),
						// master1
						createFakePod(apiserver, master1, clusterVersion),
						createFakePod(controllerManager, master1, clusterVersion),
						createFakePod(scheduler, master1, clusterVersion),
						createFakePod(etcd, master1, clusterVersion),
					},
				},
			),
			expectReturn: true,
		},
		{
			name: "a master not match cluster version",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						createFakeMasterNode(master0, clusterVersion, clusterVersion, false),
						createFakeMasterNode(master1, wrongVersion, clusterVersion, false),
					},
				},
				&corev1.PodList{
					Items: []corev1.Pod{
						// master0
						createFakePod(apiserver, master0, clusterVersion),
						createFakePod(controllerManager, master0, clusterVersion),
						createFakePod(scheduler, master0, clusterVersion),
						createFakePod(etcd, master0, clusterVersion),
						// master1
						createFakePod(apiserver, master1, wrongVersion),
						createFakePod(controllerManager, master1, clusterVersion),
						createFakePod(scheduler, master1, clusterVersion),
						createFakePod(etcd, master1, clusterVersion),
					},
				},
			),
			expectReturn: false,
		},
	}
	for _, tt := range nodes {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actualReturn, err := AllControlPlanesMatchVersion(tt.fakeClientset, cluster)
			if err != nil {
				t.Errorf("error not expected: (%v)", err)
				return
			}
			if actualReturn != tt.expectReturn {
				t.Errorf("got: (%v), want: (%v)", actualReturn, tt.expectReturn)
			}
		})
	}
}

func TestAllNodesMatchClusterVersionWithVersioningInfo(t *testing.T) {
	clusterVersion := version.MustParseSemantic("1.14.1")
	wrongVersion := version.MustParseSemantic("1.13.1")
	var nodes = []struct {
		name               string
		nodeVersionInfoMap NodeVersionInfoMap
		expectReturn       bool
	}{
		{
			name: "all nodes match cluster version",
			nodeVersionInfoMap: NodeVersionInfoMap{
				master0: {
					Nodename:         master0,
					KubeletVersion:   clusterVersion,
					APIServerVersion: clusterVersion,
				},
				master1: {
					Nodename:         master1,
					KubeletVersion:   clusterVersion,
					APIServerVersion: clusterVersion,
				},
				worker0: {
					Nodename:       worker0,
					KubeletVersion: clusterVersion,
				},
			},
			expectReturn: true,
		},
		{
			name: "a master not match cluster version",
			nodeVersionInfoMap: NodeVersionInfoMap{
				master0: {
					Nodename:         master0,
					APIServerVersion: clusterVersion,
					KubeletVersion:   wrongVersion,
				},
				master1: {
					Nodename:         master1,
					APIServerVersion: clusterVersion,
					KubeletVersion:   clusterVersion,
				},
				worker0: {
					Nodename:       worker0,
					KubeletVersion: clusterVersion,
				},
			},
			expectReturn: false,
		},
		{
			name: "a worker node not match cluster version",
			nodeVersionInfoMap: NodeVersionInfoMap{
				master0: {
					Nodename:         master0,
					KubeletVersion:   clusterVersion,
					APIServerVersion: clusterVersion,
				},
				master1: {
					Nodename:         master1,
					KubeletVersion:   clusterVersion,
					APIServerVersion: clusterVersion,
				},
				worker0: {
					Nodename:       worker0,
					KubeletVersion: wrongVersion,
				},
			},
			expectReturn: false,
		},
	}
	for _, tt := range nodes {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actualReturn := AllNodesMatchClusterVersionWithVersioningInfo(tt.nodeVersionInfoMap, clusterVersion)
			if actualReturn != tt.expectReturn {
				t.Errorf("got: (%v), want: (%v)", actualReturn, tt.expectReturn)
			}
		})
	}
}
