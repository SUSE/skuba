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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var m1 = createNode("master1")
var m2 = createNode("master2")
var m3 = createNode("master3")
var m4 = createNode("master4")
var m5 = createNode("master5")

func createNode(name string) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceSystem,
			Labels:    map[string]string{"node-role.kubernetes.io/master": ""},
		},
	}
}

var d1R0 = createDeployment("deployment1", 0)
var d1R2 = createDeployment("deployment1", 2)
var d1R3 = createDeployment("deployment2", 3)

func createDeployment(deployment string, replica int) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment,
			Namespace: metav1.NamespaceSystem,
			Labels:    map[string]string{"caasp.suse.com/skuba-replica-ha": "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(replica); return &i }(),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "t-1",
							Image: "t-1",
						},
					},
				},
			},
		},
	}
}

var d1RequiredR2 = createDeploymentWithAffinityRequired("deployment1-withRequired", 2)
var d1RequiredR3 = createDeploymentWithAffinityRequired("deployment1-withRequired", 3)
var d2RequiredR3 = createDeploymentWithAffinityRequired("deployment2-withRequired", 3)

func createDeploymentWithAffinityRequired(deployment string, replica int) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment,
			Namespace: metav1.NamespaceSystem,
			Labels:    map[string]string{"caasp.suse.com/skuba-replica-ha": "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(replica); return &i }(),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "t-1",
							Image: "t-1",
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key: "test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

var d1PreferredR2 = createDeploymentWithAffinityPreferred("deployment1-Preferred-2", 2)
var d1PreferredR3 = createDeploymentWithAffinityPreferred("deployment1-Preferred-3", 3)
var d2PreferredR3 = createDeploymentWithAffinityPreferred("deployment2-Preferred-3", 3)

func createDeploymentWithAffinityPreferred(deployment string, replica int) appsv1.Deployment {
	return appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment,
			Namespace: metav1.NamespaceSystem,
			Labels:    map[string]string{"caasp.suse.com/skuba-replica-ha": "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(replica); return &i }(),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "t-1",
							Image: "t-1",
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: corev1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchExpressions: []metav1.LabelSelectorRequirement{
												{
													Key: "test",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
