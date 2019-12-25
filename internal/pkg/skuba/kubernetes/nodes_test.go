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
	"testing"

	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktest "k8s.io/client-go/testing"
)

var m1 = createNode("master1", true, "m1")
var m2 = createNode("master2", true, "m2")
var m3 = createNode("master3", true, "m3")
var w1 = createNode("worker1", false, "w1")
var w2 = createNode("worker2", false, "w2")
var w3 = createNode("worker3", false, "w3")
var invalid = createNode("invalid", false, "invalid")

func addEvictionSupport(fakeClientset *fake.Clientset) {
	podsEviction := metav1.APIResource{
		Name:    "pods/eviction",
		Kind:    "Eviction",
		Group:   "",
		Version: "v1",
	}
	coreResources := &metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{podsEviction},
	}

	policyResources := &metav1.APIResourceList{
		GroupVersion: "policy/v1",
	}
	fakeClientset.Resources = append(fakeClientset.Resources, coreResources, policyResources)
}

func createNode(name string, isControlPlane bool, machineID string) corev1.Node {
	if isControlPlane {
		return corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: map[string]string{"node-role.kubernetes.io/master": ""},
			},
			Status: corev1.NodeStatus{
				NodeInfo: corev1.NodeSystemInfo{
					MachineID: machineID,
				},
			},
		}
	}

	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				MachineID: machineID,
			},
		},
	}
}

func TestGetAllNodes(t *testing.T) {
	tests := []struct {
		name          string
		fakeClientset *fake.Clientset
		expect        int
	}{
		{
			name: "get nodes when cluster has 1 master",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1,
					},
				},
			),
			expect: 1,
		},
		{
			name: "get control plane node when cluster has 1 master, 1 worker",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1,
						w1,
					},
				},
			),
			expect: 2,
		},
		{
			name: "get all node when cluster has 2 master, 2 worker",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2,
						w1, w2,
					},
				},
			),
			expect: 4,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actual, _ := GetAllNodes(tt.fakeClientset)
			actualSize := len(actual.Items)
			if actualSize != tt.expect {
				t.Errorf("returned node number (%d) does not match the expected one (%d)", actualSize, tt.expect)
				return
			}
		})
	}
}

func TestGetControlPlaneNodes(t *testing.T) {
	tests := []struct {
		name          string
		fakeClientset *fake.Clientset
		expect        int
	}{
		{
			name: "get control plane node when cluster has 1 master",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1,
					},
				},
			),
			expect: 1,
		},
		{
			name: "get control plane node when cluster has 2 master",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2,
					},
				},
			),
			expect: 2,
		},
		{
			name: "get control plane node when cluster has 3 master",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3,
					},
				},
			),
			expect: 3,
		},
		{
			name: "get control plane node when cluster has 3 master, 1 worker",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3,
						w1,
					},
				},
			),
			expect: 3,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actual, _ := GetControlPlaneNodes(tt.fakeClientset)
			actualSize := len(actual.Items)
			if actualSize != tt.expect {
				t.Errorf("returned master node number (%d) does not match the expected one (%d)", actualSize, tt.expect)
				return
			}
		})
	}
}

func TestGetNodeWithMachineID(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.NodeList{
			Items: []corev1.Node{
				m1, m2, m3,
				w1, w2, w3,
			},
		},
	)

	tests := []struct {
		name      string
		machineID string
		expect    string
		expectErr bool
	}{
		{
			name:      "get node name matching master node machine ID",
			machineID: m1.Status.NodeInfo.MachineID,
			expect:    m1.Name,
		},
		{
			name:      "get node name matching worker node machine ID",
			machineID: w2.Status.NodeInfo.MachineID,
			expect:    w2.Name,
		},
		{
			name:      "get node name when machine ID does not exist",
			machineID: invalid.Status.NodeInfo.MachineID,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actual, actualErr := GetNodeWithMachineID(fakeClientset, tt.machineID)
			if tt.expectErr {
				expectErr := fmt.Sprintf("node with machine-id %s not found", tt.machineID)
				if actualErr.Error() != expectErr {
					t.Errorf("returned error (%v) does not match the expected one (%v)", actualErr, expectErr)
					return
				}
				return
			}
			if actual.Name != tt.expect {
				t.Errorf("returned node name (%s) does not match the expected on (%s)", actual.Name, tt.expect)
				return
			}
		})
	}
}

func TestDrainNode(t *testing.T) {
	tests := []struct {
		name         string
		node         string
		reactorPod   string
		evictSupport bool
		expectErrMsg string
	}{
		{
			name:         "drain master node by eviction",
			node:         m1.Name,
			evictSupport: true,
			reactorPod:   "valid",
		},
		{
			name:         "drain worker node by eviction",
			node:         w1.Name,
			reactorPod:   "valid",
			evictSupport: true,
		},
		{
			name:       "drain worker node by deletion",
			node:       w1.Name,
			reactorPod: "valid",
		},
		{
			name:         "drain when node does not exist",
			node:         invalid.Name,
			evictSupport: true,
			expectErrMsg: fmt.Sprintf("failed to update node status: nodes \"%s\" not found", invalid.Name),
		},
		{
			name:         "drain by eviction when pod does not exist",
			node:         m1.Name,
			evictSupport: true,
			reactorPod:   "invalid",
			expectErrMsg: "failed to evict pod: invalid: pods \"invalid\" not found",
		},
		{
			name:         "drain by deletion when pod does not exist",
			node:         m1.Name,
			reactorPod:   "invalid",
			expectErrMsg: "failed to delete pod: invalid: pods \"invalid\" not found",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1,
						w1,
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "valid",
						Namespace: metav1.NamespaceDefault,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test",
								Image: "registry.suse.com/caasp/v4/test:1.1.1",
							},
						},
					},
				},
			)
			fakeClientset.PrependReactor("list", "pods", func(action ktest.Action) (bool, runtime.Object, error) {
				obj := &corev1.PodList{
					Items: []corev1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      tt.reactorPod,
								Namespace: metav1.NamespaceDefault,
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "test",
										Image: "registry.suse.com/caasp/v4/test:1.1.1",
									},
								},
							},
						},
					},
				}
				return true, obj, nil
			})

			if tt.evictSupport {
				addEvictionSupport(fakeClientset)
			}

			node := corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.node,
				},
			}
			err := DrainNode(fakeClientset, &node, 10)

			if tt.expectErrMsg != "" {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
					return
				}
				if err.Error() != tt.expectErrMsg {
					t.Errorf("returned error (%v) does not match the expected one (%v)", err.Error(), tt.expectErrMsg)
					return
				}
				return
			}
			if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err.Error())
				return
			}

			if tt.evictSupport {
				var actualEviction []policyv1beta1.Eviction
				for _, action := range fakeClientset.Actions() {
					if action.GetVerb() != "create" ||
						action.GetResource().Resource != "pods" ||
						action.GetSubresource() != "eviction" {
						continue
					}

					eviction := *action.(ktest.CreateAction).GetObject().(*policyv1beta1.Eviction)
					actualEviction = append(actualEviction, eviction)
				}
				if len(actualEviction) == 0 {
					t.Errorf("eviction expected on %s, but no eviction reported", tt.name)
				}
				return
			}
		})
	}
}
