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
		clientset     *fake.Clientset
		ciliumVersion string
		etcdDir       string
		errExpected   bool
		errMessage    string
		name          string
	}{
		{
			name:          "should create cilium secret",
			clientset:     fake.NewSimpleClientset(),
			ciliumVersion: "1.6",
			etcdDir:       "testdata/valid_cert_valid_key",
			errExpected:   false,
		},
		{
			name:          "should fail with invalid secret key",
			clientset:     fake.NewSimpleClientset(),
			ciliumVersion: "1.6",
			etcdDir:       "testdata/valid_cert_invalid_key",
			errExpected:   true,
			errMessage:    "etcd generation retrieval failed failed to load key: couldn't load the private key file testdata/valid_cert_invalid_key/ca.key: error reading private key file testdata/valid_cert_invalid_key/ca.key: data does not contain a valid RSA or ECDSA private key",
		},
		{
			name:          "should fail with invalid secret cert",
			clientset:     fake.NewSimpleClientset(),
			ciliumVersion: "1.6",
			etcdDir:       "testdata/invalid_cert_valid_key",
			errExpected:   true,
			errMessage:    "etcd generation retrieval failed failed to load certificate: couldn't load the certificate file testdata/invalid_cert_valid_key/ca.crt: error reading testdata/invalid_cert_valid_key/ca.crt: data does not contain any valid RSA or ECDSA certificates",
		},
		{
			name:          "should fail with invalid secret directory",
			clientset:     fake.NewSimpleClientset(),
			ciliumVersion: "1.6",
			etcdDir:       "testdata/not_exist",
			errExpected:   true,
			errMessage:    "etcd generation retrieval failed failed to load certificate: couldn't load the certificate file testdata/not_exist/ca.crt: open testdata/not_exist/ca.crt: no such file or directory",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			etcdDir = tt.etcdDir

			err := CreateCiliumSecret(tt.clientset, tt.ciliumVersion)
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
	// cmKubeadm is a fake kubeadm config map which is used to generate the
	// etcd config. Test cases which do not include that config map and
	// require including etcd config in Cilium config map should trigger an
	// error.
	cmKubeadm := &corev1.ConfigMap{
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

	// cmOldCiliumWithEtcd is a fake config map of the previous Cilium
	// deployment which contains the 'etcd-config' field, which should
	// trigger including etcd config in the first instance of updated Cilium
	// config map to perform the migration from etcd to CRD afterwards.
	// That config map is representing the case when the upgrade from Cilium
	// 1.5 to 1.6 is performed.
	cmOldCiliumWithEtcd := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ciliumConfigMapName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			"hello":       "world",
			"etcd-config": "etcd-config-content",
		},
	}

	// cmOldCiliumWithoutEtcd is a fake config map of the previous Cilium
	// deployment which does not contain etcd config, so the updated Cilium
	// config map should not include it as well and migration from etcd to
	// CRD should not be performed.
	// That config map is representing the case when the previous instance
	// is already supporting Cilium 1.6 with CRD.
	cmOldCiliumWithoutEtcd := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ciliumConfigMapName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{"hello": "world"},
	}

	tests := []struct {
		clientset                      *fake.Clientset
		ciliumVersion                  string
		kubeadmConfigMap               *corev1.ConfigMap
		ciliumConfigMap                *corev1.ConfigMap
		bpfCtGlobalTcpMaxExpected      string
		bpfCtGlobalAnyMaxExpected      string
		debugExpected                  string
		enableIpv4Expected             string
		enableIpv6Expected             string
		etcdConfigExpected             string
		identityAllocationModeExpected string
		kvstoreExpected                string
		kvstoreOptExpected             string
		preallocateBpfMapsExpected     string
		dataExpected                   map[string]string
		errExpected                    bool
		errMessage                     string
		name                           string
	}{
		// This test case represents the deployment of Cilium 1.5 when
		// Kubernetes < 1.17 is deployed.
		{
			name:             "should create or update cilium configmap for 1.5",
			clientset:        fake.NewSimpleClientset(),
			kubeadmConfigMap: cmKubeadm,
			ciliumVersion:    "1.5.3",
			dataExpected: map[string]string{
				"debug":       "false",
				"enable-ipv4": "true",
				"enable-ipv6": "false",
				"etcd-config": `ca-file: /tmp/cilium-etcd/ca.crt
cert-file: /tmp/cilium-etcd/tls.crt
endpoints:
- https://1.2.3.4:2379
key-file: /tmp/cilium-etcd/tls.key
`},
			errExpected: false,
		},
		// This test case represents the deployment of Cilium 1.5 when
		// Kubernetes < 1.17 is deployed, but the error is expected due
		// to lack of kubeadm configmap which is needed for generating
		// etcd config.
		{
			name:          "should fail when kubeadm-config configmap not exist",
			clientset:     fake.NewSimpleClientset(),
			ciliumVersion: "1.5.3",
			errExpected:   true,
			errMessage:    "unable to get api endpoints: could not retrieve the kubeadm-config configmap to get apiEndpoints: configmaps \"kubeadm-config\" not found",
		},
		// This test case represents the new deployment with Cilium 1.6
		// which should use CRD for identity allocation.
		{
			name:          "should create cilium configmap for a new 1.6 deployment",
			clientset:     fake.NewSimpleClientset(),
			ciliumVersion: "1.6.6",
			dataExpected: map[string]string{
				"bpf-ct-global-tcp-max":    "524288",
				"bpf-ct-global-any-max":    "262144",
				"debug":                    "false",
				"enable-ipv4":              "true",
				"enable-ipv6":              "false",
				"identity-allocation-mode": "crd",
				"preallocate-bpf-maps":     "false",
			},
			errExpected: false,
		},
		// This test case represents the update from Cilium 1.5 to 1.6,
		// where the updated config map should still contain etcd config
		// to be able to perform migration from etcd to CRD.
		{
			name:             "should update cilium configmap with etcd config (upgrade from 1.5 to 1.6)",
			clientset:        fake.NewSimpleClientset(),
			kubeadmConfigMap: cmKubeadm,
			ciliumConfigMap:  cmOldCiliumWithEtcd,
			ciliumVersion:    "1.6.6",
			dataExpected: map[string]string{
				"bpf-ct-global-tcp-max": "524288",
				"bpf-ct-global-any-max": "262144",
				"debug":                 "false",
				"enable-ipv4":           "true",
				"enable-ipv6":           "false",
				"etcd-config": `cert-file: /var/lib/etcd-secrets/etcd-client.crt
endpoints:
- https://1.2.3.4:2379
key-file: /var/lib/etcd-secrets/etcd-client.key
trusted-ca-file: /var/lib/etcd-secrets/etcd-client-ca.crt
`,
				"identity-allocation-mode": "kvstore",
				"kvstore":                  "etcd",
				"kvstore-opt":              "{\"etcd.config\": \"/var/lib/etcd-config/etcd.config\"}",
				"preallocate-bpf-maps":     "false",
			},
			errExpected: false,
		},
		// This test case represents the update from Cilium 1.5 to 1.6,
		// but the error is expected due to lack of kubeadm configmap
		// which is needed for generating etcd config.
		{
			name:            "should fail when kubeadm-config configmap not exist (upgrade from 1.5 to 1.6)",
			clientset:       fake.NewSimpleClientset(),
			ciliumConfigMap: cmOldCiliumWithEtcd,
			ciliumVersion:   "1.6.6",
			errExpected:     true,
			errMessage:      "unable to get api endpoints: could not retrieve the kubeadm-config configmap to get apiEndpoints: configmaps \"kubeadm-config\" not found",
		},
		// This test case represents the update of Cilium config map
		// when the previous instance is already supportin Cilium 1.6 with CRD.
		{
			name:            "should update cilium configmap without etcd config (previous configmap already for 1.6)",
			clientset:       fake.NewSimpleClientset(),
			ciliumConfigMap: cmOldCiliumWithoutEtcd,
			ciliumVersion:   "1.6.6",
			dataExpected: map[string]string{
				"bpf-ct-global-tcp-max":    "524288",
				"bpf-ct-global-any-max":    "262144",
				"debug":                    "false",
				"enable-ipv4":              "true",
				"enable-ipv6":              "false",
				"identity-allocation-mode": "crd",
				"preallocate-bpf-maps":     "false",
			},
			errExpected: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			if tt.kubeadmConfigMap != nil {
				//nolint:errcheck
				tt.clientset.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(tt.kubeadmConfigMap)
			}
			if tt.ciliumConfigMap != nil {
				//nolint:errcheck
				tt.clientset.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(tt.ciliumConfigMap)
			}

			err := CreateOrUpdateCiliumConfigMap(tt.clientset, tt.ciliumVersion)

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

				if !reflect.DeepEqual(dataGet.Data, tt.dataExpected) {
					t.Errorf("returned data (%v) does not match the expected one (%v)", dataGet.Data, tt.dataExpected)
					return
				}
			}
		})
	}
}
