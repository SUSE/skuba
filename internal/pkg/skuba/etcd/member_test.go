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

package etcd_test

import (
	"crypto/sha1"
	"fmt"
	"testing"

	"github.com/SUSE/skuba/internal/pkg/skuba/etcd"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"k8s.io/client-go/kubernetes/fake"
)

func Test_RemoveMember(t *testing.T) {
	fakeCfgFiles := map[string][]byte{
		"KubeletComponentConfig": []byte(`
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
`),
		"ClusterConfiguration": []byte(`
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
kubernetesVersion: v1.13.0
apiServer:
  extraArgs:
    advertiseAddress: 1.2.3.4
`),
	}

	fakeConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeadm-config",
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			kubeadmconstants.ClusterConfigurationConfigMapKey:     string(fakeCfgFiles["ClusterConfiguration"]),
			kubeadmconstants.KubeletBaseConfigurationConfigMapKey: string(fakeCfgFiles["KubeletComponentConfig"]),
		},
	}

	fakeDaemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "app/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceSystem,
			Name:      "cilium",
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "fake",
							Image: "fake",
						},
					},
				},
			},
		},
	}

	fakeJobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    "fake",
						Image:   "fake",
						Command: []string{"/bin/bash"},
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
				Volumes:       []corev1.Volume{},
				NodeSelector: map[string]string{
					"kubernetes.io/hostname": "fake",
				},
				Tolerations: []corev1.Toleration{
					{
						Operator: corev1.TolerationOpExists,
					},
				},
			},
		},
	}

	fakeWorker := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker",
		},
	}

	fakeMaster := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "master",
			Labels: map[string]string{"node-role.kubernetes.io/master": ""},
		},
	}

	tests := []struct {
		errExpected bool
		errMessage  string
		name        string
	}{
		{
			name:        "should remove etcd member from etcd cluster",
			errExpected: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{fakeMaster, fakeWorker}})

			clientset.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(fakeConfigMap)
			clientset.AppsV1().DaemonSets(metav1.NamespaceSystem).Create(fakeDaemonSet)

			controlPlaneNodes, _ := kubernetes.GetControlPlaneNodes(clientset)
			hashTarget := fmt.Sprintf("%x", sha1.Sum([]byte(fakeWorker.ObjectMeta.Name)))
			hashExecutor := fmt.Sprintf("%x", sha1.Sum([]byte(controlPlaneNodes.Items[0].ObjectMeta.Name)))
			job := fmt.Sprintf("caasp-remove-etcd-member-%.10s-from-%.10s", hashTarget, hashExecutor)
			clientset.BatchV1().Jobs(metav1.NamespaceSystem).Create(&batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      job,
					Namespace: metav1.NamespaceSystem,
				},
				Spec: fakeJobSpec,
			})

			controlPlaneComponentsVersion, _ := kubeadm.GetCurrentClusterVersion(clientset)

			err := etcd.RemoveMember(clientset, &fakeWorker, controlPlaneComponentsVersion)
			if tt.errExpected {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
					return
				}
				if err.Error() != tt.errMessage {
					t.Errorf("returned error (%v) does not match the expected one (%v)", err.Error(), tt.errMessage)
					return
				}
			} else if !tt.errExpected {
				if err != nil {
					t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err.Error())
					return
				}
			}
		})
	}
}
