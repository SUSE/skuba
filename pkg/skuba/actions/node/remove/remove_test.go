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

package remove

import (
	"crypto/sha1"
	"fmt"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_RemoveMasterNode(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeadm-config",
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{"ClusterConfiguration": `
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
kubernetesVersion: "v1.2.3"
apiServer:
  extraArgs:
    advertiseAddress: 1.2.3.4
`},
	}

	jobSpec := batchv1.JobSpec{
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

	master1 := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "master-1",
			Labels: map[string]string{"node-role.kubernetes.io/master": ""},
		},
	}

	master2 := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "master-2",
			Labels: map[string]string{"node-role.kubernetes.io/master": ""},
		},
	}

	worker1 := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker-1",
		},
	}

	worker2 := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker-2",
		},
	}

	test := []struct {
		name          string
		target        string
		job           string
		clientset     *fake.Clientset
		errorExpected bool
		errorMessage  string
	}{
		{
			name:          "should remove master from cluster",
			target:        "master-1",
			clientset:     fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{master1, master2}}),
			job:           fmt.Sprintf("caasp-kubelet-disarm-%x", sha1.Sum([]byte(master1.ObjectMeta.Name))),
			errorExpected: false,
		},
		{
			name:          "should fail when remove last master from cluster",
			target:        "master-1",
			clientset:     fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{master1}}),
			errorExpected: true,
			errorMessage:  "could not remove last master of the cluster",
		},
		{
			name:          "should fail when remove node does not exist",
			target:        "not-exist",
			clientset:     fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{master1}}),
			errorExpected: true,
			errorMessage:  "[remove-node] could not get node not-exist: nodes \"not-exist\" not found",
		},
		{
			name:          "should remove worker from cluster",
			target:        "worker-2",
			clientset:     fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{worker1, worker2}}),
			job:           fmt.Sprintf("caasp-kubelet-disarm-%x", sha1.Sum([]byte(worker2.ObjectMeta.Name))),
			errorExpected: false,
		},
	}

	for _, tt := range test {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			tt.clientset.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(cm)
			tt.clientset.BatchV1().Jobs(metav1.NamespaceSystem).Create(&batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.job,
					Namespace: metav1.NamespaceSystem,
				},
				Spec: jobSpec,
			})

			err := Remove(tt.clientset, tt.target, 0)
			if tt.errorExpected && err == nil {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if !tt.errorExpected && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%s)", tt.name, err.Error())
				return
			} else if tt.errorExpected && err.Error() != tt.errorMessage {
				t.Errorf("returned error (%v) does not match the expected one (%v)", err.Error(), tt.errorMessage)
				return
			}
		})
	}
}
