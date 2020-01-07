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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_getPodContainerImageTag(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
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

	tests := []struct {
		name         string
		namespace    string
		podName      string
		expect       string
		expectErrMsg string
	}{
		{
			name:      "get pod container image tag",
			namespace: metav1.NamespaceDefault,
			podName:   "valid",
			expect:    "1.1.1",
		},
		{
			name:         "get pod container image tag with pod not exist",
			namespace:    metav1.NamespaceDefault,
			podName:      "invalid",
			expectErrMsg: "could not retrieve pod object: pods \"invalid\" not found",
		},
		{
			name:         "get pod container image tag with pod exist in different namespace",
			namespace:    metav1.NamespaceSystem,
			podName:      "invalid",
			expectErrMsg: "could not retrieve pod object: pods \"invalid\" not found",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actual, err := getPodContainerImageTag(fakeClientset, tt.namespace, tt.podName)
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
			if !reflect.DeepEqual(actual, tt.expect) {
				t.Errorf("returned image tag (%v) does not match the expected one (%v)", actual, tt.expect)
				return
			}
		})
	}
}

func Test_getPodFromPodList(t *testing.T) {
	podList := newPodList(1000)
	validPod := newPod("valid")
	podList.Items = append(podList.Items, *validPod)

	tests := []struct {
		list         *corev1.PodList
		name         string
		expect       *corev1.Pod
		expectErrMsg string
	}{
		{
			list:   podList,
			name:   "valid",
			expect: validPod,
		},
		{
			list:         podList,
			name:         "invalid",
			expectErrMsg: "could not find pod invalid in pod list",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actual, err := getPodFromPodList(tt.list, tt.name)
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
			if !reflect.DeepEqual(actual, tt.expect) {
				t.Errorf("returned pod (%v) does not match the expected one (%v)", actual, tt.expect)
				return
			}
		})
	}
}

func benchmarkGetPodFromPodList(i int, b *testing.B) {
	podList := newPodList(i)
	validPod := newPod("valid")
	podList.Items = append(podList.Items, *validPod)
	for n := 0; n < b.N; n++ {
		_, err := getPodFromPodList(podList, "valid")
		if err != nil {
			b.Errorf("an error was reported (%v)", err.Error())
		}
	}
}

func Benchmark_getPodFromPodList10(b *testing.B)    { benchmarkGetPodFromPodList(10, b) }
func Benchmark_getPodFromPodList100(b *testing.B)   { benchmarkGetPodFromPodList(100, b) }
func Benchmark_getPodFromPodList1000(b *testing.B)  { benchmarkGetPodFromPodList(1000, b) }
func Benchmark_getPodFromPodList10000(b *testing.B) { benchmarkGetPodFromPodList(10000, b) }

// returns a pod object with the name passed as parameter
func newPod(name string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       corev1.PodSpec{},
		Status:     corev1.PodStatus{},
	}
}

// returns a list of pods with the number of elements passed as paramenter
func newPodList(podsNumber int) *corev1.PodList {
	pods := []corev1.Pod{}
	for i := 0; i < podsNumber; i++ {
		pods = append(pods, *newPod(fmt.Sprintf("test-pod%d", i)))
	}
	return &corev1.PodList{Items: pods}
}
