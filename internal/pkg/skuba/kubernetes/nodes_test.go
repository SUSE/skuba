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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNodeVersioningInfoWithClientset(t *testing.T) {
	testK8sVersion := version.MustParseSemantic(fmt.Sprintf("v%s", CurrentVersion))
	testEtcdVersion := version.MustParseSemantic(currentETCDVersion)
	namespace := "kube-system"
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
			containerRuntimeVersion:  fmt.Sprintf("cri-o://%s", CurrentVersion),
			kubeletVersion:           testK8sVersion,
			apiServerVersion:         testK8sVersion,
			controllerManagerVersion: testK8sVersion,
			schedulerVersion:         testK8sVersion,
			etcdVersion:              testEtcdVersion,
			expectedNodeVersionInfo: NodeVersionInfo{
				ContainerRuntimeVersion:  fmt.Sprintf("cri-o://%s", CurrentVersion),
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
			containerRuntimeVersion:  fmt.Sprintf("cri-o://%s", CurrentVersion),
			kubeletVersion:           testK8sVersion,
			apiServerVersion:         testK8sVersion,
			controllerManagerVersion: testK8sVersion,
			schedulerVersion:         testK8sVersion,
			etcdVersion:              testEtcdVersion,
			expectedNodeVersionInfo: NodeVersionInfo{
				ContainerRuntimeVersion:  fmt.Sprintf("cri-o://%s", CurrentVersion),
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
							Image: fmt.Sprintf("%s/%s/%s/hyperkube:%s", registry, productName, caaspVersion, CurrentVersion),
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
							Image: fmt.Sprintf("%s/%s/%s/hyperkube:%s", registry, productName, caaspVersion, CurrentVersion),
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
							Image: fmt.Sprintf("%s/%s/%s/hyperkube:%s", registry, productName, caaspVersion, CurrentVersion),
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
							Image: fmt.Sprintf("%s/%s/%s/etcd:%s", registry, productName, caaspVersion, currentETCDVersion),
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

func TestGetPodContainerImageTagWithClientset(t *testing.T) {
	podName := "etcd-my-master-0"
	namespace := "kube-system"
	expectedEtcdContainerImageTag := currentETCDVersion
	t.Run("get pod container image tag with clientset", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podName,
				Namespace: namespace,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "etcd",
						Image: fmt.Sprintf("%s/%s/%s/etcd:%s", registry, productName, caaspVersion, currentETCDVersion),
					},
				},
			},
		})

		etcdContainerImageTag, _ := getPodContainerImageTagWithClientset(clientset, namespace, podName)
		if !reflect.DeepEqual(etcdContainerImageTag, expectedEtcdContainerImageTag) {
			t.Errorf("got %v, want %v", etcdContainerImageTag, expectedEtcdContainerImageTag)
		}
	})
}
