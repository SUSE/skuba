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
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"

	"github.com/SUSE/skuba/pkg/skuba"
)

type componentTestData struct {
	name           string
	clusterVersion *version.Version
	component      Component
	imageName      string
	expectVersion  string
	expectErr      bool
}

type addonTestData struct {
	name                  string
	clusterVersion        *version.Version
	addon                 Addon
	expectVersion         string
	expectManifestVersion uint
	expectErr             bool
}

type versionTestData struct {
	hostComponent      []componentTestData
	containerComponent []componentTestData
	addon              []addonTestData
}

func getComponentTestData(component Component, version *version.Version, info *KubernetesVersion) componentTestData {
	var name string
	var expectVersion string
	var imageName string
	switch component {
	case Kubelet:
		name = fmt.Sprintf("get %s version when cluster version is %s", component, version)
		expectVersion = info.ComponentHostVersion.KubeletVersion
	case ContainerRuntime:
		name = fmt.Sprintf("get %s version when cluster version is %s", component, version)
		expectVersion = info.ComponentHostVersion.ContainerRuntimeVersion
	case Hyperkube, Etcd, CoreDNS, Pause, Tooling:
		name = fmt.Sprintf("get %s image when cluster version is %s", component, version)
		imageName = info.ComponentContainerVersion[component].Name
		expectVersion = info.ComponentContainerVersion[component].Tag
	}
	return componentTestData{
		name:           name,
		clusterVersion: version,
		component:      component,
		imageName:      imageName,
		expectVersion:  expectVersion,
	}
}

func getAddonTestData(addon Addon, version *version.Version, info *KubernetesVersion) addonTestData {
	return addonTestData{
		name:                  fmt.Sprintf("get %s version when cluster version is %s", addon, version),
		clusterVersion:        version,
		addon:                 addon,
		expectVersion:         info.AddonsVersion[addon].Version,
		expectManifestVersion: info.AddonsVersion[addon].ManifestVersion,
	}
}

func getVersionTestData() versionTestData {
	var addonTestData []addonTestData
	var clusterVersion *version.Version
	var containerComponentTestData []componentTestData
	var hostComponentTestData []componentTestData
	for ver, verInfo := range supportedVersions {
		clusterVersion = version.MustParseSemantic(ver)
		hostComponentTestData = append(
			hostComponentTestData,
			getComponentTestData(Kubelet, clusterVersion, &verInfo),
			getComponentTestData(ContainerRuntime, clusterVersion, &verInfo),
		)

		for cc := range verInfo.ComponentContainerVersion {
			containerComponentTestData = append(
				containerComponentTestData,
				getComponentTestData(cc, clusterVersion, &verInfo),
			)
		}

		for addon := range verInfo.AddonsVersion {
			addonTestData = append(
				addonTestData,
				getAddonTestData(addon, clusterVersion, &verInfo),
			)
		}
	}

	return versionTestData{
		hostComponent:      hostComponentTestData,
		containerComponent: containerComponentTestData,
		addon:              addonTestData,
	}
}

var vtd = getVersionTestData()

func TestComponentVersionForClusterVersion(t *testing.T) {
	var tests = vtd.hostComponent
	tests = append(
		tests,
		componentTestData{
			name:           "invalid component",
			clusterVersion: version.MustParseSemantic("1.16.2"),
			component:      "not-exist",
			expectVersion:  "",
		},
	)

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actualVersion := ComponentVersionForClusterVersion(tt.component, tt.clusterVersion)
			if actualVersion != tt.expectVersion {
				t.Errorf("returned version (%s) does not match the expected one (%s)", actualVersion, tt.expectVersion)
				return
			}
		})
	}
}

func TestAllComponentContainerImagesForClusterVersion(t *testing.T) {
	var clusterVersion *version.Version
	for ver := range supportedVersions {
		clusterVersion = version.MustParseSemantic(ver)
		t.Run(fmt.Sprintf("get all component container images when cluster version is %s", ver), func(t *testing.T) {
			actual := AllComponentContainerImagesForClusterVersion(clusterVersion)
			sort.Slice(actual, func(i int, j int) bool {
				return actual[i] < actual[j]
			})
			expect := []Component{Hyperkube, Etcd, CoreDNS, Pause, Tooling}
			sort.Slice(expect, func(i int, j int) bool {
				return expect[i] < expect[j]
			})

			if !reflect.DeepEqual(actual, expect) {
				t.Errorf("returned result (%s) does not match the expected one (%s)", actual, expect)
				return
			}
		})
	}
}

func TestComponentContainerImageForClusterVersion(t *testing.T) {
	var tests = vtd.containerComponent
	tests = append(
		tests,
		componentTestData{
			name:           "invalid component",
			clusterVersion: version.MustParseSemantic("1.16.2"),
			component:      "not-exist",
			expectErr:      true,
		},
	)

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actual := ComponentContainerImageForClusterVersion(tt.component, tt.clusterVersion)
			if tt.expectErr {
				if actual != "" {
					t.Errorf("image not expected. but a result was returned (%s)", actual)
					return
				}
			} else {
				expect := images.GetGenericImage(skuba.ImageRepository, tt.imageName, tt.expectVersion)
				if actual != expect {
					t.Errorf("returned image (%s) does not match the expected one (%s)", actual, expect)
					return
				}
			}
		})
	}
}

func TestAddonVersionForClusterVersion(t *testing.T) {
	var tests = vtd.addon
	tests = append(
		tests,
		addonTestData{
			name:           "invalid addon",
			clusterVersion: version.MustParseSemantic("1.16.2"),
			addon:          "not-exist",
			expectErr:      true,
		},
	)

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actual := AddonVersionForClusterVersion(tt.addon, tt.clusterVersion)
			if tt.expectErr {
				if actual != nil {
					t.Errorf("addon version not expected. but a result was returned (%v)", *actual)
					return
				}
			} else {
				if actual.Version != tt.expectVersion {
					t.Errorf("returned addon version (%s) does not match the expected one (%s)", actual.Version, tt.expectVersion)
					return
				}
				if actual.ManifestVersion != tt.expectManifestVersion {
					t.Errorf("returned addon version (%v) does not match the expected one (%v)", actual.ManifestVersion, tt.expectManifestVersion)
					return
				}
			}
		})
	}
}

func TestAllAddonVersionsForClusterVersion(t *testing.T) {
	type test struct {
		name           string
		clusterVersion *version.Version
	}
	var tests []test
	for ver := range supportedVersions {
		tests = append(
			tests,
			test{
				name:           fmt.Sprintf("get all addon versions when cluster version is %s", ver),
				clusterVersion: version.MustParseSemantic(ver),
			},
		)
	}
	tests = append(
		tests,
		test{
			name:           "invalid cluster version",
			clusterVersion: version.MustParseSemantic("0.0.0"),
		},
	)

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actual := AllAddonVersionsForClusterVersion(tt.clusterVersion)
			expect := supportedVersions[tt.clusterVersion.String()].AddonsVersion

			if !reflect.DeepEqual(actual, expect) {
				actualData, err := json.Marshal(actual)
				if err != nil {
					t.Errorf("error not expected while convert returned result to json data (%v)", err)
					return
				}
				expectData, err := json.Marshal(expect)
				if err != nil {
					t.Errorf("error not expected while convert returned result to json data (%v)", err)
					return
				}
				t.Errorf("returned result (%s) does not match the expected one (%s)", actualData, expectData)
				return
			}
		})
	}
}

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
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			availableVersions := AvailableVersionsForMap(tt.kubernetesVersions)
			if !reflect.DeepEqual(availableVersions, tt.expectedAvailableVersions) {
				t.Errorf("got %q, want %q", availableVersions, tt.expectedAvailableVersions)
			}
		})
	}
}

func TestLatestVersion(t *testing.T) {
	if _, ok := supportedVersions[LatestVersion().String()]; !ok {
		t.Errorf("Versions map --authoritative version mapping-- does not include version %q", LatestVersion().String())
	}
}

func TestIsVersionAvailable(t *testing.T) {
	if !IsVersionAvailable(LatestVersion()) {
		t.Errorf("Versions map does not include version %q", LatestVersion().String())
	}
}

func TestMajorMinorVersion(t *testing.T) {
	tests := []struct {
		name                      string
		version                   *version.Version
		expectedMajorMinorVersion string
	}{
		{
			name:                      "without prefix v",
			version:                   version.MustParseSemantic("1.14.1"),
			expectedMajorMinorVersion: "1.14",
		},
		{
			name:                      "with prefix v",
			version:                   version.MustParseSemantic("v1.14.1"),
			expectedMajorMinorVersion: "1.14",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotVersion := MajorMinorVersion(tt.version)
			if gotVersion != tt.expectedMajorMinorVersion {
				t.Errorf("got version %s, expect version %s", gotVersion, tt.expectedMajorMinorVersion)
			}
		})
	}
}
