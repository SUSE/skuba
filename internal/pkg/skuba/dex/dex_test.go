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

package dex

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_GenerateClientSecret(t *testing.T) {
	tests := []struct {
		name      string
		inputLen  int
		expectLen int
	}{
		{
			name:      "normal case",
			expectLen: 24,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateClientSecret()
			gotLen := len(got)

			// Compare length
			if gotLen != tt.expectLen {
				t.Errorf("got len %d != expect len %d", gotLen, tt.expectLen)
			}

			// check string is all 0 or not
			if strings.Trim(got, "0") == "" {
				t.Error("got is not randomly")
			}
		})
	}
}

func Test_CreateCert(t *testing.T) {
	tests := []struct {
		name                string
		pkiPath             string
		kubeadmInitConfPath string
		expectedError       bool
	}{
		{
			name:                "normal case",
			pkiPath:             "testdata",
			kubeadmInitConfPath: "testdata/kubeadm-init.conf",
		},
		{
			name:                "invalid pki path",
			pkiPath:             "invalid-pki-path",
			kubeadmInitConfPath: "testdata/kubeadm-init.conf",
			expectedError:       true,
		},
		{
			name:                "invalid kubeadm init path",
			pkiPath:             "testdata",
			kubeadmInitConfPath: "testdata/invalid-kubeadm-init-conf-path.conf",
			expectedError:       true,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			err := CreateCert(fake.NewSimpleClientset(), tt.pkiPath, tt.kubeadmInitConfPath)
			if tt.expectedError && err == nil {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if !tt.expectedError && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
		})
	}
}

func Test_DexCertExists(t *testing.T) {
	tests := []struct {
		name          string
		client        clientset.Interface
		expectedExist bool
		expectedError bool
	}{
		{
			name: "certificate exists",
			client: fake.NewSimpleClientset(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "oidc-dex-cert",
					Namespace: metav1.NamespaceSystem,
				},
			}),
			expectedExist: true,
			expectedError: false,
		},
		{
			name:          "certificate not exists",
			client:        fake.NewSimpleClientset(),
			expectedExist: false,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := DexCertExists(tt.client)
			if tt.expectedError {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
				}
				return
			} else if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}

			if got != tt.expectedExist {
				t.Errorf("expect %t, got %t\n", tt.expectedExist, got)
			}
		})
	}
}

func Test_RestartPods(t *testing.T) {
	tests := []struct {
		name          string
		client        clientset.Interface
		expectedError bool
	}{
		{
			name: "restart pod successfully",
			client: fake.NewSimpleClientset(&corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "dex-pod-1",
							Labels: map[string]string{"app": "oidc-dex"},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "dex-pod-2",
							Labels: map[string]string{"app": "oidc-dex"},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:   "dex-pod-3",
							Labels: map[string]string{"app": "oidc-dex"},
						},
					},
				},
			}),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := RestartPods(tt.client)
			if tt.expectedError {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
				}
				return
			} else if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
		})
	}
}
