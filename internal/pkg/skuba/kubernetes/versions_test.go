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

	"k8s.io/apimachinery/pkg/util/version"
)

func TestAvailableVersionsForMap(t *testing.T) {
	var versions = []struct {
		name                      string
		kubernetesVersions        KubernetesVersions
		expectedAvailableVersions []*version.Version
	}{
		{
			name: "v1.14.0-v1.15.0",
			kubernetesVersions: KubernetesVersions{
				"v1.14.0": KubernetesVersion{},
				"v1.15.0": KubernetesVersion{},
			},
			expectedAvailableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.0"),
				version.MustParseSemantic("v1.15.0"),
			},
		},
		{
			name: "v1.14.0-v1.14.1-v1.15.0",
			kubernetesVersions: KubernetesVersions{
				"v1.14.0": KubernetesVersion{},
				"v1.15.0": KubernetesVersion{},
				"v1.14.1": KubernetesVersion{},
			},
			expectedAvailableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.0"),
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.15.0"),
			},
		},
	}
	for _, tt := range versions {
		t.Run(tt.name, func(t *testing.T) {
			availableVersions := availableVersionsForMap(tt.kubernetesVersions)
			if !reflect.DeepEqual(availableVersions, tt.expectedAvailableVersions) {
				t.Errorf("got %q, want %q", availableVersions, tt.expectedAvailableVersions)
			}
		})
	}
}

func TestLatestVersion(t *testing.T) {
	if _, ok := Versions[LatestVersion().String()]; !ok {
		t.Errorf("Versions map --authoritative version mapping-- does not include version %q", LatestVersion().String())
	}
}

func TestCurrentComponentVersion(t *testing.T) {
	components := []Component{Hyperkube, Etcd, CoreDNS, Pause}
	for _, component := range components {
		t.Run(fmt.Sprintf("component %q has a version assigned", component), func(t *testing.T) {
			componentVersion := CurrentComponentVersion(component)
			_, err := version.ParseGeneric(componentVersion)
			if err != nil {
				t.Errorf("component %q version (%q) parsing failed", component, componentVersion)
			}
		})
	}
}
