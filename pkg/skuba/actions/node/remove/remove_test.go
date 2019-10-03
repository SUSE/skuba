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

func Test_RemoveNode(t *testing.T) {
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
		executer      []string
		clientset     *fake.Clientset
		errorExpected bool
		errorMessage  string
	}{
		{
			name:          "should remove master from cluster",
			target:        master2.Name,
			executer:      []string{master1.Name, master2.Name},
			clientset:     fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{master1, master2}}),
			errorExpected: false,
		},
		{
			name:          "should fail when remove last master from cluster",
			target:        master1.Name,
			clientset:     fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{master1}}),
			errorExpected: true,
			errorMessage:  "could not remove last master of the cluster",
		},
		{
			name:          "should fail when remove node does not exist",
			target:        "not-exist",
			executer:      []string{master1.Name},
			clientset:     fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{master1}}),
			errorExpected: true,
			errorMessage:  "[remove-node] could not get node not-exist: nodes \"not-exist\" not found",
		},
		{
			name:          "should remove worker from cluster",
			target:        worker2.Name,
			executer:      []string{master1.Name},
			clientset:     fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{master1, worker1, worker2}}),
			errorExpected: false,
		},
	}

	for _, tt := range test {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			tt.clientset.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(cm)

			shaTarget := fmt.Sprintf("%x", sha1.Sum([]byte(tt.target)))
			shaExecuter := ""
			for _, executer := range tt.executer {
				shaExecuter = fmt.Sprintf("%x", sha1.Sum([]byte(executer)))
				tt.clientset.BatchV1().Jobs(metav1.NamespaceSystem).Create(&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("caasp-remove-etcd-member-%.10s-from-%.10s", shaTarget, shaExecuter),
						Namespace: metav1.NamespaceSystem,
					},
					Spec:   jobSpec,
					Status: batchv1.JobStatus{Active: 1},
				})
			}

			tt.clientset.BatchV1().Jobs(metav1.NamespaceSystem).Create(&batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("caasp-kubelet-disarm-%s", shaTarget),
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
