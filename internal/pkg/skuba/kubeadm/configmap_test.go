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

package kubeadm

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes/fake"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

var kubeadmConfigMap = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      kubeadmconstants.KubeadmConfigConfigMap,
		Namespace: metav1.NamespaceSystem,
	},
	Data: map[string]string{
		"ClusterConfiguration": `
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
kubernetesVersion: v1.16.2
`,
		"ClusterStatus": `
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterStatus
apiEndpoints:
  master1:
    advertiseAddress: 192.168.0.1
  master2:
    advertiseAddress: 192.168.0.2
  master3:
    advertiseAddress: 192.168.0.3
`},
}

func TestGetCurrentClusterVersion(t *testing.T) {
	tests := []struct {
		name            string
		client          *fake.Clientset
		expectedError   bool
		expectedVersion *version.Version
	}{
		{
			name:          "kubeadm configmap not found",
			client:        fake.NewSimpleClientset(),
			expectedError: true,
		},
		{
			name: "kubeadm configmap no ClusterConfiguration",
			client: fake.NewSimpleClientset(
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kubeadmconstants.KubeadmConfigConfigMap,
						Namespace: metav1.NamespaceSystem,
					},
				},
			),
			expectedError: true,
		},
		{
			name: "kubeadm configmap decode failed",
			client: fake.NewSimpleClientset(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kubeadmconstants.KubeadmConfigConfigMap,
					Namespace: metav1.NamespaceSystem,
				},
				Data: map[string]string{"ClusterConfiguration": `
					apiVersion: kubeadm.k8s.io/v1beta2
					kind: ClusterConfiguration
				`},
			}),
			expectedError: true,
		},
		{
			name:            "kubernetes cluster version v1.16.2",
			client:          fake.NewSimpleClientset(kubeadmConfigMap),
			expectedVersion: version.MustParseSemantic("v1.16.2"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, err := GetCurrentClusterVersion(tt.client)
			if tt.expectedError {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
				}
				return
			}

			if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
			if !reflect.DeepEqual(gotVersion, tt.expectedVersion) {
				t.Errorf("got version %v, expect version %v", gotVersion, tt.expectedVersion)
			}
		})
	}
}

func TestGetKubeadmApisVersion(t *testing.T) {
	versions := []struct {
		version                  string
		expectKubeadmApisVersion string
	}{
		{
			version:                  "1.14.2",
			expectKubeadmApisVersion: "v1beta1",
		},
		{
			version:                  "1.15.0",
			expectKubeadmApisVersion: "v1beta2",
		},
		{
			version:                  "1.15.2",
			expectKubeadmApisVersion: "v1beta2",
		},
		{
			version:                  "1.16.2",
			expectKubeadmApisVersion: "v1beta2",
		},
	}

	for _, tt := range versions {
		tt := tt
		t.Run(tt.version, func(t *testing.T) {
			gotKubeadmApisVersion := GetKubeadmApisVersion(version.MustParseSemantic(tt.version))
			if gotKubeadmApisVersion != tt.expectKubeadmApisVersion {
				t.Errorf("got kubeadm api version %s does not match expected kubeadm api version %s", gotKubeadmApisVersion, tt.expectKubeadmApisVersion)
			}
		})
	}
}

func TestGetAPIEndpointsFromConfigMap(t *testing.T) {
	tests := []struct {
		name                 string
		client               *fake.Clientset
		expectedError        bool
		expectedAPIEndpoints []string
	}{
		{
			name:          "kubeadm configmap not found",
			client:        fake.NewSimpleClientset(),
			expectedError: true,
		},
		{
			name: "kubeadm configmap no ClusterStatus",
			client: fake.NewSimpleClientset(
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kubeadmconstants.KubeadmConfigConfigMap,
						Namespace: metav1.NamespaceSystem,
					},
				},
			),
			expectedError: true,
		},
		{
			name: "kubeadm configmap decode failed",
			client: fake.NewSimpleClientset(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kubeadmconstants.KubeadmConfigConfigMap,
					Namespace: metav1.NamespaceSystem,
				},
				Data: map[string]string{"ClusterStatus": `
					apiVersion: kubeadm.k8s.io/v1beta2
					kind: ClusterStatus
				`},
			}),
			expectedError: true,
		},
		{
			name:                 "kubeadm get api endpoints from configmap",
			client:               fake.NewSimpleClientset(kubeadmConfigMap),
			expectedAPIEndpoints: []string{"192.168.0.1", "192.168.0.2", "192.168.0.3"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotAPIEndpoints, err := GetAPIEndpointsFromConfigMap(tt.client)
			if tt.expectedError {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
				}
				return
			}

			if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
			if !reflect.DeepEqual(gotAPIEndpoints, tt.expectedAPIEndpoints) {
				t.Errorf("got version %v, expect version %v", gotAPIEndpoints, tt.expectedAPIEndpoints)
			}
		})
	}
}

func TestRemoveAPIEndpointFromConfigMap(t *testing.T) {
	tests := []struct {
		name                 string
		client               *fake.Clientset
		node                 *corev1.Node
		expectedError        bool
		expectedAPIEndpoints []string
	}{
		{
			name:          "kubeadm configmap not found",
			client:        fake.NewSimpleClientset(),
			expectedError: true,
		},
		{
			name: "kubeadm configmap no ClusterStatus",
			client: fake.NewSimpleClientset(
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kubeadmconstants.KubeadmConfigConfigMap,
						Namespace: metav1.NamespaceSystem,
					},
				},
			),
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "master1",
					Labels: map[string]string{"node-role.kubernetes.io/master": ""},
				},
			},
			expectedAPIEndpoints: []string{},
		},
		{
			name: "kubeadm configmap decode failed",
			client: fake.NewSimpleClientset(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kubeadmconstants.KubeadmConfigConfigMap,
					Namespace: metav1.NamespaceSystem,
				},
				Data: map[string]string{"ClusterStatus": `
					apiVersion: kubeadm.k8s.io/v1beta2
					kind: ClusterStatus
				`},
			}),
			expectedError: true,
		},
		{
			name:   "remove master node",
			client: fake.NewSimpleClientset(kubeadmConfigMap),
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "master1",
					Labels: map[string]string{"node-role.kubernetes.io/master": ""},
				},
			},
			expectedAPIEndpoints: []string{"192.168.0.2", "192.168.0.3"},
		},
		{
			name:   "remove worker node",
			client: fake.NewSimpleClientset(kubeadmConfigMap),
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker1",
				},
			},
			expectedAPIEndpoints: []string{"192.168.0.1", "192.168.0.2", "192.168.0.3"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := RemoveAPIEndpointFromConfigMap(tt.client, tt.node)
			if tt.expectedError {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
				}
				return
			}

			if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}

			gotAPIEndpoints, _ := GetAPIEndpointsFromConfigMap(tt.client)

			sort.Slice(gotAPIEndpoints, func(i, j int) bool {
				return gotAPIEndpoints[i] < gotAPIEndpoints[j]
			})
			sort.Slice(tt.expectedAPIEndpoints, func(i, j int) bool {
				return tt.expectedAPIEndpoints[i] < tt.expectedAPIEndpoints[j]
			})

			if !reflect.DeepEqual(gotAPIEndpoints, tt.expectedAPIEndpoints) {
				t.Errorf("got version %v, expect version %v", gotAPIEndpoints, tt.expectedAPIEndpoints)
			}
		})
	}
}

func TestUpdateClusterConfigurationWithClusterVersion(t *testing.T) {
	var scenarios = []struct {
		name                     string
		clusterVersion           *version.Version
		currentAdmissionPlugins  []string
		expectedAdmissionPlugins []string
	}{
		{
			name:                     "1.15.2 without duplicates",
			clusterVersion:           version.MustParseSemantic("1.15.2"),
			currentAdmissionPlugins:  []string{},
			expectedAdmissionPlugins: []string{"NamespaceLifecycle", "LimitRanger", "ServiceAccount", "TaintNodesByCondition", "Priority", "DefaultTolerationSeconds", "DefaultStorageClass", "PersistentVolumeClaimResize", "MutatingAdmissionWebhook", "ValidatingAdmissionWebhook", "ResourceQuota", "StorageObjectInUseProtection", "NodeRestriction", "PodSecurityPolicy"},
		},
		{
			name:                     "1.15.2 with duplicates",
			clusterVersion:           version.MustParseSemantic("1.15.2"),
			currentAdmissionPlugins:  []string{"NamespaceLifecycle", "NodeRestriction", "PodSecurityPolicy"},
			expectedAdmissionPlugins: []string{"NamespaceLifecycle", "NodeRestriction", "PodSecurityPolicy", "LimitRanger", "ServiceAccount", "TaintNodesByCondition", "Priority", "DefaultTolerationSeconds", "DefaultStorageClass", "PersistentVolumeClaimResize", "MutatingAdmissionWebhook", "ValidatingAdmissionWebhook", "ResourceQuota", "StorageObjectInUseProtection"},
		},
		{
			name:                     "1.16.2 without duplicates",
			clusterVersion:           version.MustParseSemantic("1.16.2"),
			currentAdmissionPlugins:  []string{},
			expectedAdmissionPlugins: []string{"NamespaceLifecycle", "LimitRanger", "ServiceAccount", "TaintNodesByCondition", "Priority", "DefaultTolerationSeconds", "DefaultStorageClass", "PersistentVolumeClaimResize", "MutatingAdmissionWebhook", "ValidatingAdmissionWebhook", "ResourceQuota", "StorageObjectInUseProtection", "RuntimeClass", "NodeRestriction", "PodSecurityPolicy"},
		},
	}

	for _, tt := range scenarios {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			expectedAdmissionPlugins := strings.Join(tt.expectedAdmissionPlugins, ",")
			initCfg := kubeadmapi.InitConfiguration{}
			if len(tt.currentAdmissionPlugins) > 0 {
				currentAdmissionPlugins := strings.Join(tt.currentAdmissionPlugins, ",")
				if initCfg.APIServer.ControlPlaneComponent.ExtraArgs == nil {
					initCfg.APIServer.ControlPlaneComponent.ExtraArgs = map[string]string{}
				}
				initCfg.APIServer.ControlPlaneComponent.ExtraArgs["enable-admission-plugins"] = currentAdmissionPlugins
			}
			UpdateClusterConfigurationWithClusterVersion(&initCfg, tt.clusterVersion)
			// Check admission plugins
			gotAdmissionPlugins := initCfg.APIServer.ControlPlaneComponent.ExtraArgs["enable-admission-plugins"]
			if gotAdmissionPlugins != expectedAdmissionPlugins {
				t.Errorf("admission plugins %s do not match expected admission plugins %s", gotAdmissionPlugins, expectedAdmissionPlugins)
			}
			// Check different configuration settings
			if initCfg.ImageRepository != skuba.ImageRepository {
				t.Errorf("image repository %s does not match expected image repository %s", initCfg.ImageRepository, skuba.ImageRepository)
			}
			expectedClusterVersion := fmt.Sprintf("v%s", tt.clusterVersion.String())
			if initCfg.KubernetesVersion != expectedClusterVersion {
				t.Errorf("kubernetes version %s does not match expected kubernetes version %s", initCfg.KubernetesVersion, expectedClusterVersion)
			}
			if initCfg.Etcd.Local.ImageRepository != skuba.ImageRepository {
				t.Errorf("etcd image repository %s does not match expected etcd image repository %s", initCfg.Etcd.Local.ImageRepository, skuba.ImageRepository)
			}
			etcdExpectedTag := kubernetes.ComponentVersionForClusterVersion(kubernetes.Etcd, tt.clusterVersion)
			if initCfg.Etcd.Local.ImageTag != etcdExpectedTag {
				t.Errorf("etcd image tag %s does not match expected etcd image tag %s", initCfg.Etcd.Local.ImageTag, etcdExpectedTag)
			}
			coreDNSExpectedTag := kubernetes.ComponentVersionForClusterVersion(kubernetes.CoreDNS, tt.clusterVersion)
			if initCfg.DNS.ImageTag != coreDNSExpectedTag {
				t.Errorf("coredns image tag %s does not match expected coredns image tag %s", initCfg.DNS.ImageTag, coreDNSExpectedTag)
			}
		})
	}
}
