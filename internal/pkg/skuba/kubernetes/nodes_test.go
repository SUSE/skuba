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
)

func TestNodeVersioningInfoWithClientset(t *testing.T) {
	var nodes = []struct {
		name                    string
		nodeName                string
		unschedulable           bool
		kubeletVersion          *version.Version
		containerRuntimeVersion string
		expectedNodeVersionInfo NodeVersionInfo
	}{
		{
			name:                    "node version info schedulable",
			nodeName:                "my-worker-0",
			unschedulable:           false,
			containerRuntimeVersion: "cri-o://1.14.1",
			kubeletVersion:          version.MustParseSemantic("v1.14.1"),
			expectedNodeVersionInfo: NodeVersionInfo{
				ContainerRuntimeVersion: "cri-o://1.14.1",
				KubeletVersion:          version.MustParseSemantic("v1.14.1"),
				Unschedulable:           false,
			},
		},
		{
			name:                    "node version info unschedulable",
			nodeName:                "my-worker-0",
			unschedulable:           true,
			containerRuntimeVersion: "cri-o://1.14.1",
			kubeletVersion:          version.MustParseSemantic("v1.14.1"),
			expectedNodeVersionInfo: NodeVersionInfo{
				ContainerRuntimeVersion: "cri-o://1.14.1",
				KubeletVersion:          version.MustParseSemantic("v1.14.1"),
				Unschedulable:           true,
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
			})

			nodeVersionInfo, _ := nodeVersioningInfoWithClientset(clientset, tt.nodeName)
			if !reflect.DeepEqual(nodeVersionInfo, tt.expectedNodeVersionInfo) {
				t.Errorf("got %v, want %v", nodeVersionInfo, tt.expectedNodeVersionInfo)
			}
		})
	}
}
