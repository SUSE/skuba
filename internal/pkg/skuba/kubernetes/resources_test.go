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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_DoesResourceExistWithError(t *testing.T) {
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

	fakeClientset := fake.NewSimpleClientset(
		&corev1.NodeList{
			Items: []corev1.Node{
				fakeMaster,
				fakeWorker,
			},
		},
		&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceSystem,
				Name:      "test",
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"apiKey": []byte("test"),
			},
		},
	)

	tests := []struct {
		name             string
		searchNamespace  string
		searchSecret     string
		expectReturn     bool
		expectErrMessage error
	}{
		{
			name:             "secret exist with no error",
			searchNamespace:  metav1.NamespaceSystem,
			searchSecret:     "test",
			expectReturn:     true,
			expectErrMessage: nil,
		},
		{
			name:             "secret not exist",
			searchNamespace:  metav1.NamespaceSystem,
			searchSecret:     "not-exist",
			expectReturn:     false,
			expectErrMessage: nil,
		},
		{
			name:             "secret exist in other namespace",
			searchNamespace:  metav1.NamespaceDefault,
			searchSecret:     "test",
			expectReturn:     false,
			expectErrMessage: nil,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			_, err := fakeClientset.CoreV1().Secrets(tt.searchNamespace).Get(tt.searchSecret, metav1.GetOptions{})
			actualReturn, actualErrMessage := DoesResourceExistWithError(err)
			if actualReturn != tt.expectReturn {
				t.Errorf("returned (%v) does not match the expected one (%v)", actualReturn, tt.expectReturn)
				return
			}
			if actualErrMessage != tt.expectErrMessage {
				t.Errorf("returned error (%v) does not match the expected one (%v)", actualErrMessage, tt.expectErrMessage)
				return
			}
		})
	}
}
