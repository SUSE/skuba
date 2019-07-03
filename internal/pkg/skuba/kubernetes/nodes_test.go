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
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetPodContainerImageTagWithClientset(t *testing.T) {
	podName := "etcd-my-master-0"
	namespace := "kube-system"
	expectedEtcdContainerImageTag := "3.3.11"
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
						Image: "registry.suse.com/caasp/v4/etcd:3.3.11",
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
