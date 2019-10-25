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

package cni

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/fake"
)

func Test_CreateCiliumSecret(t *testing.T) {
	tests := []struct {
		clientset   *fake.Clientset
		etcdDir     string
		errExpected bool
		errMessage  string
		name        string
	}{
		{
			name:        "should create cilium secret",
			clientset:   fake.NewSimpleClientset(),
			etcdDir:     "testdata/valid_cert_valid_key",
			errExpected: false,
		},
		{
			name:        "should fail with invalid secret key",
			clientset:   fake.NewSimpleClientset(),
			etcdDir:     "testdata/valid_cert_invalid_key",
			errExpected: true,
			errMessage:  "etcd generation retrieval failed failed to load key: couldn't load the private key file testdata/valid_cert_invalid_key/ca.key: error reading private key file testdata/valid_cert_invalid_key/ca.key: data does not contain a valid RSA or ECDSA private key",
		},
		{
			name:        "should fail with invalid secret cert",
			clientset:   fake.NewSimpleClientset(),
			etcdDir:     "testdata/invalid_cert_valid_key",
			errExpected: true,
			errMessage:  "etcd generation retrieval failed failed to load certificate: couldn't load the certificate file testdata/invalid_cert_valid_key/ca.crt: error reading testdata/invalid_cert_valid_key/ca.crt: data does not contain any valid RSA or ECDSA certificates",
		},
		{
			name:        "should fail with invalid secret directory",
			clientset:   fake.NewSimpleClientset(),
			etcdDir:     "testdata/not_exist",
			errExpected: true,
			errMessage:  "etcd generation retrieval failed failed to load certificate: couldn't load the certificate file testdata/not_exist/ca.crt: open testdata/not_exist/ca.crt: no such file or directory",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			etcdDir = tt.etcdDir

			err := CreateCiliumSecret(tt.clientset)
			//nolint:errcheck
			secrets, _ := tt.clientset.CoreV1().Secrets(metav1.NamespaceSystem).List(metav1.ListOptions{})
			secretSize := len(secrets.Items)
			if tt.errExpected {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
					return
				}
				if err.Error() != tt.errMessage {
					t.Errorf("returned error (%v) does not match the expected one (%v)", err.Error(), tt.errMessage)
					return
				}
				if secretSize != 0 {
					t.Errorf("secret not expected on %s, but %d secret was found", tt.name, secretSize)
					return
				}
			} else if !tt.errExpected {
				if err != nil {
					t.Errorf("error not expected on %s, but an error was reported (%s)", tt.name, err.Error())
					return
				}
				if secretSize == 0 {
					t.Errorf("secret expected on %s, but no secret was found", tt.name)
					return
				}
			}
		})
	}
}

func Test_AnnotateCiliumDaemonsetWithCurrentTimestamp(t *testing.T) {
	dsBase := &appsv1.DaemonSet{
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

	dsNotExist := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "app/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceSystem,
			Name:      "invalid",
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

	tests := []struct {
		clientset   *fake.Clientset
		daemonset   *appsv1.DaemonSet
		errExpected bool
		errMessage  string
		name        string
	}{
		{
			name:        "should annotate cilium daemonset with current timestamp",
			clientset:   fake.NewSimpleClientset(),
			daemonset:   dsBase,
			errExpected: false,
		},
		{
			name:        "should fail when daemonset does not exist",
			clientset:   fake.NewSimpleClientset(),
			daemonset:   dsNotExist,
			errExpected: true,
			errMessage:  "daemonsets.apps \"cilium\" not found",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			//nolint:errcheck
			tt.clientset.AppsV1().DaemonSets(metav1.NamespaceSystem).Create(tt.daemonset)

			err := annotateCiliumDaemonsetWithCurrentTimestamp(tt.clientset)

			if tt.errExpected {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
					return
				}
				if err.Error() != tt.errMessage {
					t.Errorf("returned error (%v) does not match the expected one (%v)", err.Error(), tt.errMessage)
					return
				}
			} else if !tt.errExpected && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%s)", tt.name, err.Error())
				return
			}
		})
	}
}

func Test_CreateOrUpdateCiliumConfigMap(t *testing.T) {
	cmBase := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeadm-config",
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{"ClusterStatus": `
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterStatus
apiEndpoints:
  master:
    advertiseAddress: 1.2.3.4
    bindPort: 80
`},
	}

	cmDefaultNamespace := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake",
			Namespace: metav1.NamespaceDefault,
		},
		Data: map[string]string{"hello": "world"},
	}

	tests := []struct {
		clientset          *fake.Clientset
		configmap          *corev1.ConfigMap
		debugExpected      string
		enableIpv4Expected string
		enableIpv6Expected string
		etcdConfigExpected string
		errExpected        bool
		errMessage         string
		name               string
	}{
		{
			name:               "should create or update cilium configmap",
			clientset:          fake.NewSimpleClientset(),
			configmap:          cmBase,
			debugExpected:      "false",
			enableIpv4Expected: "true",
			enableIpv6Expected: "false",
			errExpected:        false,
			etcdConfigExpected: `ca-file: /tmp/cilium-etcd/ca.crt
cert-file: /tmp/cilium-etcd/tls.crt
endpoints:
- https://1.2.3.4:2379
key-file: /tmp/cilium-etcd/tls.key
`},
		{
			name:        "should fail when kubeadm-config configmap not exist",
			clientset:   fake.NewSimpleClientset(),
			configmap:   cmDefaultNamespace,
			errExpected: true,
			errMessage:  "unable to get api endpoints: could not retrieve the kubeadm-config configmap to get apiEndpoints: configmaps \"kubeadm-config\" not found",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			//nolint:errcheck
			tt.clientset.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(tt.configmap)

			err := CreateOrUpdateCiliumConfigMap(tt.clientset)

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

				dataGet, err := tt.clientset.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get("cilium-config", metav1.GetOptions{})
				if err != nil {
					t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err.Error())
					return
				}

				dataExpected := map[string]string{
					"debug":       tt.debugExpected,
					"enable-ipv4": tt.enableIpv4Expected,
					"enable-ipv6": tt.enableIpv6Expected,
					"etcd-config": tt.etcdConfigExpected,
				}
				if !reflect.DeepEqual(dataGet.Data, dataExpected) {
					t.Errorf("returned data (%v) does not match the expected one (%v)", dataGet.Data, dataExpected)
					return
				}
			}
		})
	}
}
