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
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
)

func TestUpdatedAddons(t *testing.T) {
	type test struct {
		name           string
		client         clientset.Interface
		clusterVersion *version.Version
		expectedAviu   AddonVersionInfoUpdate
		expectedErr    bool
	}

	tests := []test{}
	tests = append(
		tests,
		test{
			name:           "no skuba-config ConfigMap",
			client:         fake.NewSimpleClientset(),
			clusterVersion: version.MustParseSemantic("1.2.3"),
			expectedAviu: AddonVersionInfoUpdate{
				Current: kubernetes.AddonsVersion{},
				Updated: kubernetes.AddonsVersion{},
			},
		},
		test{
			name: "skuba-config format error",
			client: fake.NewSimpleClientset(
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      skuba.ConfigMapName,
						Namespace: metav1.NamespaceSystem,
					},
					Data: map[string]string{
						skuba.SkubaConfigurationKeyName: `
							AddonsVersion:
						`,
					},
				},
			),
			clusterVersion: version.MustParseSemantic("1.2.3"),
			expectedAviu: AddonVersionInfoUpdate{
				Current: kubernetes.AddonsVersion{},
				Updated: kubernetes.AddonsVersion{},
			},
			expectedErr: true,
		},
	)

	// Test without updated
	for _, cv := range kubernetes.AvailableVersions() {
		client := fake.NewSimpleClientset()
		aviu := AddonVersionInfoUpdate{
			Current: kubernetes.AddonsVersion{},
			Updated: kubernetes.AddonsVersion{},
		}
		avs := kubernetes.AllAddonVersionsForClusterVersion(cv)

		// Update skuba-config configmap
		err := skuba.UpdateSkubaConfiguration(client, &skuba.SkubaConfiguration{AddonsVersion: avs})
		if err != nil {
			t.Errorf("error not expected but an error was reported %v", err)
			return
		}
		for addon, v := range avs {
			aviu.Current[addon] = v
		}

		tests = append(
			tests,
			test{
				name:           fmt.Sprintf("kubernetes version %s without updated", cv.String()),
				client:         client,
				clusterVersion: cv,
				expectedAviu:   aviu,
			},
		)
	}

	// Test with updated
	for _, cv := range kubernetes.AvailableVersions() {
		client := fake.NewSimpleClientset()
		aviu := AddonVersionInfoUpdate{
			Current: kubernetes.AddonsVersion{},
			Updated: kubernetes.AddonsVersion{},
		}
		avs := kubernetes.AllAddonVersionsForClusterVersion(cv)
		for addon, v := range avs {
			aviu.Current[addon] = nil
			aviu.Updated[addon] = v
		}

		tests = append(
			tests,
			test{
				name:           fmt.Sprintf("kubernetes version %s with updated", cv.String()),
				client:         client,
				clusterVersion: cv,
				expectedAviu:   aviu,
			},
		)
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			gotAviu, err := UpdatedAddons(tt.client, tt.clusterVersion)
			if tt.expectedErr {
				if err == nil {
					t.Error("error expected but no error reported")
				}
				return
			}

			if err != nil {
				t.Errorf("error not expected but an error was reported %v", err)
				return
			}
			if !reflect.DeepEqual(gotAviu, tt.expectedAviu) {
				t.Errorf("got: %v, expect: %v", gotAviu, tt.expectedAviu)
				return
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
				return
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
}
