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

package replica

import (
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

func TestNewHelper(t *testing.T) {
	t.Run("create new helper", func(t *testing.T) {
		fakeClientset := fake.NewSimpleClientset(
			&corev1.NodeList{
				Items: []corev1.Node{
					m1, m2,
				},
			},
			&appsv1.DeploymentList{
				Items: []appsv1.Deployment{
					d1R2,
				},
			},
		)

		newReplica, err := NewHelper(fakeClientset)
		if err != nil {
			t.Errorf("error not expected, but an error was reported (%v)", err.Error())
		}

		nodes, err := kubernetes.GetAllNodes(fakeClientset)
		if err != nil {
			t.Errorf("error not expected, but an error was reported (%v)", err.Error())
		}
		expect := len(nodes.Items)
		if newReplica.ClusterSize != expect {
			t.Errorf("returned cluster size (%v) does not match the expected one (%v)", newReplica.ClusterSize, expect)
		}

		deployments, err := fakeClientset.AppsV1().Deployments(metav1.NamespaceSystem).List(
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=true", highAvailabilitylabel),
			},
		)
		if err != nil {
			t.Errorf("error not expected, but an error was reported (%v)", err.Error())
		}
		actual := len(newReplica.SelectDeployments.Items)
		expect = len(deployments.Items)
		if actual != expect {
			t.Errorf("returned deployment size (%v) does not match the expected one (%v)", actual, expect)
		}
	})
}

func TestUpdateNodes(t *testing.T) {
	tests := []struct {
		name          string
		fakeClientset *fake.Clientset
	}{
		{
			name: "2nd master nodes joins cluster, deployments: [replica-0]",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1R0,
					},
				},
			),
		},
		{
			name: "2nd master nodes joins cluster, deployments: [replica-2]",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1R2,
					},
				},
			),
		},
		{
			name: "2nd master node joins cluster, deployments: [replica-2, replica-3]",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1R2, d1R3,
					},
				},
			),
		},
		{
			name: "3rd master node joins cluster, deployments: [replica-2]",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1RequiredR2,
					},
				},
			),
		},
		{
			name: "3rd master node joins cluster, deployments: [replica-2, replica-3]",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1RequiredR2, d1PreferredR3,
					},
				},
			),
		},
		{
			name: "4th master node joins cluster",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3, m4,
					},
				},
			),
		},
		{
			name: "5th master node joins cluster",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3, m4, m5,
					},
				},
			),
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			replicaHelper, _ := NewHelper(tt.fakeClientset)
			err := replicaHelper.UpdateNodes()
			if err != nil {
				t.Errorf("error not expected when (%s), but an error was reported (%v)", tt.name, err.Error())
			}

			deployments, _ := tt.fakeClientset.AppsV1().Deployments(metav1.NamespaceSystem).List(
				metav1.ListOptions{
					LabelSelector: fmt.Sprintf("%s=true", highAvailabilitylabel),
				},
			)
			for _, d := range deployments.Items {
				replicaSize := int(*d.Spec.Replicas)
				switch nodes := replicaHelper.ClusterSize; {
				case nodes >= replicaSize:
					if len(d.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) == 0 {
						t.Error("expect affinity \"requiredDuringSchedulingIgnoredDuringExecution\", but affinity not exist")
					}
				case nodes >= replicaHelper.MinSize:
					if len(d.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
						t.Error("expect affinity \"preferredDuringSchedulingIgnoredDuringExecution\", but affinity not exist")
					}
				}
			}
		})
	}
}

func TestUpdateDrainNodes(t *testing.T) {
	tests := []struct {
		name          string
		fakeClientset *fake.Clientset
	}{
		{
			name: "5 master in cluster",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3, m4, m5,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1RequiredR3,
					},
				},
			),
		},
		{
			name: "4 master in cluster",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3, m4,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1RequiredR3,
					},
				},
			),
		},
		{
			name: "3 master in cluster, deployments: [replica-3]",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1RequiredR3,
					},
				},
			),
		},
		{
			name: "3 master in cluster, deployments: [replica-3, replica-3]",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2, m3,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1RequiredR3, d2RequiredR3,
					},
				},
			),
		},
		{
			name: "2 master in cluster, deployments: [replica-2]",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1PreferredR2,
					},
				},
			),
		},
		{
			name: "2 master in cluster, 2 labeled deployment: [replica-2, replica-3]",
			fakeClientset: fake.NewSimpleClientset(
				&corev1.NodeList{
					Items: []corev1.Node{
						m1, m2,
					},
				},
				&appsv1.DeploymentList{
					Items: []appsv1.Deployment{
						d1RequiredR2, d2PreferredR3,
					},
				},
			),
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			replicaHelper, _ := NewHelper(tt.fakeClientset)
			err := replicaHelper.UpdateDrainNodes()
			if err != nil {
				t.Errorf("error not expected when (%s), but an error was reported (%v)", tt.name, err.Error())
			}

			deployments, _ := tt.fakeClientset.AppsV1().Deployments(metav1.NamespaceSystem).List(
				metav1.ListOptions{
					LabelSelector: fmt.Sprintf("%s=true", highAvailabilitylabel),
				},
			)
			for _, d := range deployments.Items {
				replicaSize := int(*d.Spec.Replicas)
				preferredList := len(d.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
				requiredList := len(d.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
				switch nodes := replicaHelper.ClusterSize; {
				case nodes > replicaSize:
					if requiredList == 0 {
						t.Error("expect affinity \"requiredDuringSchedulingIgnoredDuringExecution\", but affinity not exist")
					}
				case nodes <= replicaSize:
					if preferredList == 0 {
						t.Error("expect affinity \"preferredDuringSchedulingIgnoredDuringExecution\" but affinity not exist")
					}
				}
			}
		})
	}
}
