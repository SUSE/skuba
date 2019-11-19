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
	"strings"
	"testing"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	"k8s.io/apimachinery/pkg/util/version"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

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
