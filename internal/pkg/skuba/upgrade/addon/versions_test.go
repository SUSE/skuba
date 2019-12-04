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

package addon

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	skubaconfig "github.com/SUSE/skuba/internal/pkg/skuba/skuba"
)

func TestUpdatedAddons(t *testing.T) {
	tests := []struct {
		name           string
		clusterVersion *version.Version
		clientSet      clientset.Interface
		expected       AddonVersionInfoUpdate
		expectedErr    bool
	}{
		{
			name:           "no skuba-config ConfigMap",
			clusterVersion: version.MustParseSemantic("1.15.1"),
			clientSet:      fake.NewSimpleClientset(),
			expected: AddonVersionInfoUpdate{
				Current: kubernetes.AddonsVersion{},
				Updated: kubernetes.AddonsVersion{},
			},
		},
		{
			name:           "kubernetes version 1.15.2",
			clusterVersion: version.MustParseSemantic("1.15.2"),
			clientSet: fake.NewSimpleClientset(&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      skubaconfig.ConfigMapName,
					Namespace: metav1.NamespaceSystem,
				},
				Data: map[string]string{
					skubaconfig.SkubaConfigurationKeyName: `
AddonsVersion:
  cilium:
    ManifestVersion: 32767
    Version: 1.5.3
  kured:
    ManifestVersion: 32767
    Version: 1.2.0
  dex:
    ManifestVersion: 32767
    Version: 2.16.0
  gangway:
    ManifestVersion: 32767
    Version: 3.1.0
`,
				},
			}),
			expected: AddonVersionInfoUpdate{
				Current: kubernetes.AddonsVersion{
					kubernetes.Cilium:  &kubernetes.AddonVersion{Version: "1.5.3", ManifestVersion: 32767},
					kubernetes.Kured:   &kubernetes.AddonVersion{Version: "1.2.0", ManifestVersion: 32767},
					kubernetes.Dex:     &kubernetes.AddonVersion{Version: "2.16.0", ManifestVersion: 32767},
					kubernetes.Gangway: &kubernetes.AddonVersion{Version: "3.1.0", ManifestVersion: 32767},
					kubernetes.PSP:     nil,
				},
				Updated: kubernetes.AddonsVersion{
					kubernetes.PSP: &kubernetes.AddonVersion{Version: "", ManifestVersion: 1},
				},
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := UpdatedAddons(tt.clientSet, tt.clusterVersion)

			if tt.expectedErr && gotErr == nil {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if !tt.expectedErr && gotErr != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, gotErr)
				return
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got: %v, expect: %v", got, tt.expected)
			}
		})
	}
}

func TestHasAddonUpdate(t *testing.T) {
	tests := []struct {
		name     string
		aviu     AddonVersionInfoUpdate
		expected bool
	}{
		{
			name: "has addon update",
			aviu: AddonVersionInfoUpdate{
				Current: kubernetes.AddonsVersion{
					kubernetes.Cilium:  &kubernetes.AddonVersion{Version: "1.5.3", ManifestVersion: 0},
					kubernetes.Kured:   &kubernetes.AddonVersion{Version: "1.2.0", ManifestVersion: 0},
					kubernetes.Dex:     &kubernetes.AddonVersion{Version: "2.16.0", ManifestVersion: 0},
					kubernetes.Gangway: &kubernetes.AddonVersion{Version: "3.1.0", ManifestVersion: 0},
					kubernetes.PSP:     &kubernetes.AddonVersion{Version: "1.0.0", ManifestVersion: 1},
				},
				Updated: kubernetes.AddonsVersion{
					kubernetes.Cilium: &kubernetes.AddonVersion{Version: "1.5.3", ManifestVersion: 1},
				},
			},
			expected: true,
		},
		{
			name: "no addon update",
			aviu: AddonVersionInfoUpdate{
				Current: kubernetes.AddonsVersion{
					kubernetes.Cilium:  &kubernetes.AddonVersion{Version: "1.5.3", ManifestVersion: 0},
					kubernetes.Kured:   &kubernetes.AddonVersion{Version: "1.2.0", ManifestVersion: 0},
					kubernetes.Dex:     &kubernetes.AddonVersion{Version: "2.16.0", ManifestVersion: 0},
					kubernetes.Gangway: &kubernetes.AddonVersion{Version: "3.1.0", ManifestVersion: 0},
					kubernetes.PSP:     &kubernetes.AddonVersion{Version: "1.0.0", ManifestVersion: 1},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			got := HasAddonUpdate(tt.aviu)
			if got != tt.expected {
				t.Errorf("got: %v, expect: %v", got, tt.expected)
			}
		})
	}
}

func ExamplePrintAddonUpdates() {
	PrintAddonUpdates(AddonVersionInfoUpdate{
		Current: kubernetes.AddonsVersion{
			kubernetes.Cilium:  &kubernetes.AddonVersion{Version: "1.5.3", ManifestVersion: 0},
			kubernetes.Kured:   &kubernetes.AddonVersion{Version: "1.2.0", ManifestVersion: 0},
			kubernetes.Dex:     &kubernetes.AddonVersion{Version: "2.16.0", ManifestVersion: 1},
			kubernetes.Gangway: nil,
		},
		Updated: kubernetes.AddonsVersion{
			kubernetes.Cilium:  &kubernetes.AddonVersion{Version: "1.5.3", ManifestVersion: 1},
			kubernetes.Dex:     &kubernetes.AddonVersion{Version: "2.17.0", ManifestVersion: 1},
			kubernetes.Gangway: &kubernetes.AddonVersion{Version: "3.1.0", ManifestVersion: 0},
		},
	})

	// Output:
	//   - cilium: 1.5.3 -> 1.5.3 (manifest version from 0 to 1)
	//   - dex: 2.16.0 -> 2.17.0
	//   - gangway: 3.1.0 (new addon)
	//
	// Please, run `skuba addon upgrade apply` in order to upgrade addons.
}
