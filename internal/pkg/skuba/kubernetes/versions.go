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
	"log"
	"sort"

	"k8s.io/apimachinery/pkg/util/version"
)

type Addon string
type Component string

const (
	Cilium  Addon = "cilium"
	Kured   Addon = "kured"
	Dex     Addon = "dex"
	Gangway Addon = "gangway"
	PSP     Addon = "psp"

	ContainerRuntime Component = "cri-o"
	Kubelet          Component = "kubelet"

	Hyperkube Component = "hyperkube"
	Etcd      Component = "etcd"
	CoreDNS   Component = "coredns"
	Pause     Component = "pause"

	Tooling Component = "tooling"
)

type ControlPlaneComponentsVersion struct {
	HyperkubeVersion string
	EtcdVersion      string
}

type ComponentsVersion struct {
	ContainerRuntimeVersion string
	KubeletVersion          string
	CoreDNSVersion          string
	PauseVersion            string
	ToolingVersion          string
}

type AddonVersion struct {
	Version         string
	ManifestVersion uint
}

type AddonsVersion map[Addon]*AddonVersion

type KubernetesVersion struct {
	ControlPlaneComponentsVersion ControlPlaneComponentsVersion
	ComponentsVersion             ComponentsVersion
	AddonsVersion                 AddonsVersion
}

type KubernetesVersions map[string]KubernetesVersion

var (
	Versions = KubernetesVersions{
		"1.15.2": KubernetesVersion{
			ControlPlaneComponentsVersion: ControlPlaneComponentsVersion{
				HyperkubeVersion: "v1.15.2",
				EtcdVersion:      "3.3.11",
			},
			ComponentsVersion: ComponentsVersion{
				KubeletVersion:          "1.15.2",
				ContainerRuntimeVersion: "1.15.0",
				CoreDNSVersion:          "1.3.1",
				PauseVersion:            "3.1",
				ToolingVersion:          "0.1.0",
			},
			AddonsVersion: AddonsVersion{
				Cilium:  &AddonVersion{"1.5.3", 1},
				Kured:   &AddonVersion{"1.2.0", 0},
				Dex:     &AddonVersion{"2.16.0", 3},
				Gangway: &AddonVersion{"3.1.0-rev4", 3},
				PSP:     &AddonVersion{"", 0},
			},
		},
		"1.15.0": KubernetesVersion{
			ControlPlaneComponentsVersion: ControlPlaneComponentsVersion{
				HyperkubeVersion: "v1.15.0",
				EtcdVersion:      "3.3.11",
			},
			ComponentsVersion: ComponentsVersion{
				KubeletVersion:          "1.15.0",
				ContainerRuntimeVersion: "1.15.0",
				CoreDNSVersion:          "1.3.1",
				PauseVersion:            "3.1",
				ToolingVersion:          "0.1.0",
			},
			AddonsVersion: AddonsVersion{
				Cilium:  &AddonVersion{"1.5.3", 1},
				Kured:   &AddonVersion{"1.2.0", 0},
				Dex:     &AddonVersion{"2.16.0", 3},
				Gangway: &AddonVersion{"3.1.0-rev4", 3},
				PSP:     &AddonVersion{"", 0},
			},
		},
		"1.14.1": KubernetesVersion{
			ControlPlaneComponentsVersion: ControlPlaneComponentsVersion{
				HyperkubeVersion: "v1.14.1",
				EtcdVersion:      "3.3.11",
			},
			ComponentsVersion: ComponentsVersion{
				ContainerRuntimeVersion: "1.14.1",
				KubeletVersion:          "1.14.1",
				CoreDNSVersion:          "1.2.6",
				PauseVersion:            "3.1",
				ToolingVersion:          "0.1.0",
			},
			AddonsVersion: AddonsVersion{
				Cilium:  &AddonVersion{"1.5.3", 1},
				Kured:   &AddonVersion{"1.2.0", 0},
				Dex:     &AddonVersion{"2.16.0", 3},
				Gangway: &AddonVersion{"3.1.0-rev4", 3},
				PSP:     &AddonVersion{"", 0},
			},
		},
	}
)

func ComponentVersionWithAvailableVersions(component Component, clusterVersion *version.Version, availableVersions KubernetesVersions) string {
	currentKubernetesVersion := availableVersions[clusterVersion.String()]
	switch component {
	case Hyperkube:
		return currentKubernetesVersion.ControlPlaneComponentsVersion.HyperkubeVersion
	case Etcd:
		return currentKubernetesVersion.ControlPlaneComponentsVersion.EtcdVersion
	case ContainerRuntime:
		return currentKubernetesVersion.ComponentsVersion.ContainerRuntimeVersion
	case Kubelet:
		return currentKubernetesVersion.ComponentsVersion.KubeletVersion
	case CoreDNS:
		return currentKubernetesVersion.ComponentsVersion.CoreDNSVersion
	case Pause:
		return currentKubernetesVersion.ComponentsVersion.PauseVersion
	case Tooling:
		return currentKubernetesVersion.ComponentsVersion.ToolingVersion
	}
	log.Fatalf("unknown component %q", component)
	panic("unreachable")
}

func ComponentVersionForClusterVersion(component Component, clusterVersion *version.Version) string {
	return ComponentVersionWithAvailableVersions(component, clusterVersion, Versions)
}

func AddonVersionWithAvailableVersions(addon Addon, clusterVersion *version.Version, availableVersions KubernetesVersions) *AddonVersion {
	currentKubernetesVersion := availableVersions[clusterVersion.String()]
	if addonVersion, found := currentKubernetesVersion.AddonsVersion[addon]; found {
		return addonVersion
	}
	return nil
}

func AddonVersionForClusterVersion(addon Addon, clusterVersion *version.Version) *AddonVersion {
	return AddonVersionWithAvailableVersions(addon, clusterVersion, Versions)
}

func AvailableVersionsForMap(versions KubernetesVersions) []*version.Version {
	rawVersions := make([]*version.Version, 0, len(versions))
	for rawVersion := range versions {
		rawVersions = append(rawVersions, version.MustParseSemantic(rawVersion))
	}
	sort.SliceStable(rawVersions, func(i, j int) bool {
		return rawVersions[i].LessThan(rawVersions[j])
	})
	return rawVersions
}

// AvailableVersions return the list of platform versions known to skuba
func AvailableVersions() []*version.Version {
	return AvailableVersionsForMap(Versions)
}

// LatestVersion return the latest Kubernetes supported version
func LatestVersion() *version.Version {
	availableVersions := AvailableVersions()
	return availableVersions[len(availableVersions)-1]
}

// IsVersionAvailable returns if a specific kubernetes version is available
func IsVersionAvailable(kubernetesVersion *version.Version) bool {
	_, ok := Versions[kubernetesVersion.String()]
	return ok
}

// MajorMinorVersion returns a KubernetesVersion without the patch level
func MajorMinorVersion(kubernetesVersion *version.Version) string {
	return fmt.Sprintf("%d.%d", kubernetesVersion.Major(), kubernetesVersion.Minor())
}
