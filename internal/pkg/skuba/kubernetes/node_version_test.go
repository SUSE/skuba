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
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes/fake"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

func TestNodeVersioningInfoWithClientset(t *testing.T) {
	testK8sVersion := version.MustParseSemantic("v1.14.1")
	testEtcdVersion := version.MustParseSemantic("3.3.11")
	namespace := metav1.NamespaceSystem
	var nodes = []struct {
		name                     string
		nodeName                 string
		unschedulable            bool
		kubeletVersion           *version.Version
		apiServerVersion         *version.Version
		controllerManagerVersion *version.Version
		schedulerVersion         *version.Version
		etcdVersion              *version.Version
		containerRuntimeVersion  string
		expectedNodeVersionInfo  NodeVersionInfo
	}{
		{
			name:                     "node version info schedulable",
			nodeName:                 "my-master-0",
			unschedulable:            false,
			containerRuntimeVersion:  "cri-o://1.14.1",
			kubeletVersion:           testK8sVersion,
			apiServerVersion:         testK8sVersion,
			controllerManagerVersion: testK8sVersion,
			schedulerVersion:         testK8sVersion,
			etcdVersion:              testEtcdVersion,
			expectedNodeVersionInfo: NodeVersionInfo{
				Nodename:                 "my-master-0",
				ContainerRuntimeVersion:  testK8sVersion,
				KubeletVersion:           testK8sVersion,
				APIServerVersion:         testK8sVersion,
				ControllerManagerVersion: testK8sVersion,
				SchedulerVersion:         testK8sVersion,
				EtcdVersion:              testEtcdVersion,
				Unschedulable:            false,
			},
		},
		{
			name:                     "node version info unschedulable",
			nodeName:                 "my-master-0",
			unschedulable:            true,
			containerRuntimeVersion:  "cri-o://1.14.1",
			kubeletVersion:           testK8sVersion,
			apiServerVersion:         testK8sVersion,
			controllerManagerVersion: testK8sVersion,
			schedulerVersion:         testK8sVersion,
			etcdVersion:              testEtcdVersion,
			expectedNodeVersionInfo: NodeVersionInfo{
				Nodename:                 "my-master-0",
				ContainerRuntimeVersion:  testK8sVersion,
				KubeletVersion:           testK8sVersion,
				APIServerVersion:         testK8sVersion,
				ControllerManagerVersion: testK8sVersion,
				SchedulerVersion:         testK8sVersion,
				EtcdVersion:              testEtcdVersion,
				Unschedulable:            true,
			},
		},
	}
	for _, tt := range nodes {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(&v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.nodeName,
					Labels: map[string]string{
						kubeadmconstants.LabelNodeRoleMaster: "",
					},
				},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{
						KubeletVersion:          tt.kubeletVersion.String(),
						ContainerRuntimeVersion: tt.containerRuntimeVersion,
					},
				},
				Spec: v1.NodeSpec{
					Unschedulable: tt.unschedulable,
				},
			}, &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-apiserver-my-master-0",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "kube-apiserver",
							Image: "registry.suse.com/caasp/v4/hyperkube:1.14.1",
						},
					},
					NodeName: "my-master-0",
				},
			}, &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-controller-manager-my-master-0",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "kube-controller-manager",
							Image: "registry.suse.com/caasp/v4/hyperkube:1.14.1",
						},
					},
					NodeName: "my-master-0",
				},
			}, &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-scheduler-my-master-0",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "kube-scheduler",
							Image: "registry.suse.com/caasp/v4/hyperkube:1.14.1",
						},
					},
					NodeName: "my-master-0",
				},
			}, &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "etcd-my-master-0",
					Namespace: namespace,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "etcd",
							Image: "registry.suse.com/caasp/v4/etcd:3.3.11",
						},
					},
					NodeName: "my-master-0",
				},
			})

			nodeVersionInfo, _ := nodeVersioningInfoWithClientset(clientset, tt.nodeName)
			if !reflect.DeepEqual(nodeVersionInfo, tt.expectedNodeVersionInfo) {
				t.Errorf("got %v, want %v", nodeVersionInfo, tt.expectedNodeVersionInfo)
			}
		})
	}
}

func TestAllWorkerNodesTolerateVersionWithVersioningInfo(t *testing.T) {
	var nodes = []struct {
		name                    string
		currentClusterVersion   *version.Version
		expectedResult          bool
		containerRuntimeVersion string
		nodeVersionInfoMap      NodeVersionInfoMap
	}{
		{
			name:                  "all workers tolerate",
			currentClusterVersion: version.MustParseSemantic("1.14.1"),
			expectedResult:        true,
			nodeVersionInfoMap: NodeVersionInfoMap{
				"my-master-0": {
					Nodename:                 "my-master-0",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         version.MustParseSemantic("1.14.1"),
					ControllerManagerVersion: version.MustParseSemantic("1.14.1"),
					SchedulerVersion:         version.MustParseSemantic("1.14.1"),
					EtcdVersion:              version.MustParseSemantic("3.1.11"),
					Unschedulable:            false,
				},
				"my-worker-0": {
					Nodename:                 "my-worker-0",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         nil,
					ControllerManagerVersion: nil,
					SchedulerVersion:         nil,
					EtcdVersion:              nil,
				},
				"my-worker-1": {
					Nodename:                 "my-worker-1",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         nil,
					ControllerManagerVersion: nil,
					SchedulerVersion:         nil,
					EtcdVersion:              nil,
				},
			},
		},
		{
			name:                  "a worker needs to be updated before",
			currentClusterVersion: version.MustParseSemantic("1.15.0"),
			expectedResult:        false,
			nodeVersionInfoMap: NodeVersionInfoMap{
				"my-master-0": {
					Nodename:                 "my-master-0",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.15.0"),
					KubeletVersion:           version.MustParseSemantic("1.15.0"),
					APIServerVersion:         version.MustParseSemantic("1.15.0"),
					ControllerManagerVersion: version.MustParseSemantic("1.15.0"),
					SchedulerVersion:         version.MustParseSemantic("1.15.0"),
					EtcdVersion:              version.MustParseSemantic("3.1.11"),
				},
				"my-worker-0": {
					Nodename:                 "my-worker-0",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         nil,
					ControllerManagerVersion: nil,
					SchedulerVersion:         nil,
					EtcdVersion:              nil,
				},
				"my-worker-1": {
					Nodename:                 "my-worker-1",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.13.1"),
					KubeletVersion:           version.MustParseSemantic("1.13.1"),
					APIServerVersion:         nil,
					ControllerManagerVersion: nil,
					SchedulerVersion:         nil,
					EtcdVersion:              nil,
				},
			},
		},
		{
			name:                  "all workers tolerate except an unschedulable one",
			currentClusterVersion: version.MustParseSemantic("1.15.0"),
			expectedResult:        true,
			nodeVersionInfoMap: NodeVersionInfoMap{
				"my-master-0": {
					Nodename:                 "my-master-0",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.15.0"),
					KubeletVersion:           version.MustParseSemantic("1.15.0"),
					APIServerVersion:         version.MustParseSemantic("1.15.0"),
					ControllerManagerVersion: version.MustParseSemantic("1.15.0"),
					SchedulerVersion:         version.MustParseSemantic("1.15.0"),
					EtcdVersion:              version.MustParseSemantic("3.1.11"),
				},
				"my-worker-0": {
					Nodename:                 "my-worker-0",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         nil,
					ControllerManagerVersion: nil,
					SchedulerVersion:         nil,
					EtcdVersion:              nil,
				},
				"my-worker-1": {
					Nodename:                 "my-worker-1",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.13.1"),
					KubeletVersion:           version.MustParseSemantic("1.13.1"),
					APIServerVersion:         nil,
					ControllerManagerVersion: nil,
					SchedulerVersion:         nil,
					EtcdVersion:              nil,
					Unschedulable:            true,
				},
			},
		},
	}
	for _, tt := range nodes {
		t.Run(tt.name, func(t *testing.T) {
			allWorkerNodesTolerateVersionWithVersioningInfo := allWorkerNodesTolerateVersionWithVersioningInfo(tt.nodeVersionInfoMap, tt.currentClusterVersion)
			if !reflect.DeepEqual(allWorkerNodesTolerateVersionWithVersioningInfo, tt.expectedResult) {
				t.Errorf("got %v, want %v", allWorkerNodesTolerateVersionWithVersioningInfo, tt.expectedResult)
			}
		})
	}
}

func TestAllControlPlanesMatchVersionWithVersioningInfo(t *testing.T) {
	var nodes = []struct {
		name                    string
		currentClusterVersion   *version.Version
		expectedResult          bool
		containerRuntimeVersion string
		nodeVersionInfoMap      NodeVersionInfoMap
	}{
		{
			name:                  "all control planes match",
			currentClusterVersion: version.MustParseSemantic("1.14.1"),
			expectedResult:        true,
			nodeVersionInfoMap: NodeVersionInfoMap{
				"my-master-0": {
					Nodename:                 "my-master-0",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         version.MustParseSemantic("1.14.1"),
					ControllerManagerVersion: version.MustParseSemantic("1.14.1"),
					SchedulerVersion:         version.MustParseSemantic("1.14.1"),
					EtcdVersion:              version.MustParseSemantic("3.1.11"),
					Unschedulable:            false,
				},
				"my-master-1": {
					Nodename:                 "my-master-1",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         version.MustParseSemantic("1.14.1"),
					ControllerManagerVersion: version.MustParseSemantic("1.14.1"),
					SchedulerVersion:         version.MustParseSemantic("1.14.1"),
					EtcdVersion:              version.MustParseSemantic("3.1.11"),
					Unschedulable:            false,
				},
				"my-worker-1": {
					Nodename:                 "my-worker-1",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         nil,
					ControllerManagerVersion: nil,
					SchedulerVersion:         nil,
					EtcdVersion:              nil,
					Unschedulable:            false,
				},
			},
		},
		{
			name:                  "at least one control plane doesn't match",
			currentClusterVersion: version.MustParseSemantic("1.14.1"),
			expectedResult:        false,
			nodeVersionInfoMap: NodeVersionInfoMap{
				"my-master-0": {
					Nodename:                 "my-master-0",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         version.MustParseSemantic("1.14.1"),
					ControllerManagerVersion: version.MustParseSemantic("1.14.1"),
					SchedulerVersion:         version.MustParseSemantic("1.14.1"),
					EtcdVersion:              version.MustParseSemantic("3.1.11"),
					Unschedulable:            false,
				},
				"my-master-1": {
					Nodename:                 "my-master-1",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.13.1"),
					KubeletVersion:           version.MustParseSemantic("1.13.1"),
					APIServerVersion:         version.MustParseSemantic("1.13.1"),
					ControllerManagerVersion: version.MustParseSemantic("1.13.1"),
					SchedulerVersion:         version.MustParseSemantic("1.13.1"),
					EtcdVersion:              version.MustParseSemantic("3.1.11"),
					Unschedulable:            false,
				},
				"my-worker-1": {
					Nodename:                 "my-worker-1",
					ContainerRuntimeVersion:  version.MustParseSemantic("1.14.1"),
					KubeletVersion:           version.MustParseSemantic("1.14.1"),
					APIServerVersion:         nil,
					ControllerManagerVersion: nil,
					SchedulerVersion:         nil,
					EtcdVersion:              nil,
					Unschedulable:            false,
				},
			},
		},
	}
	for _, tt := range nodes {
		t.Run(tt.name, func(t *testing.T) {
			allControlPlanesMatchVersionWithVersioningInfo := AllControlPlanesMatchVersionWithVersioningInfo(tt.nodeVersionInfoMap, tt.currentClusterVersion)
			if !reflect.DeepEqual(allControlPlanesMatchVersionWithVersioningInfo, tt.expectedResult) {
				t.Errorf("got %v, want %v", allControlPlanesMatchVersionWithVersioningInfo, tt.expectedResult)
			}
		})
	}
}
