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

// Update the latest currently supported versions
const (
	CurrentVersion        = "1.15.0"
	currentETCDVersion    = "3.3.11"
	currentCoreDNSVersion = "1.3.1"
	currentPauseVersion   = "3.1"
	currentCiliumVersion  = "1.5.3"
	currentToolingVersion = "0.1.0"
	currentKuredVersion   = "1.2.0"
	registry              = "registry.suse.com"
	productName           = "caasp"
	caaspVersion          = "v4"
)

type Addon string
type Component string

const (
	Cilium  Addon = "cilium"
	Tooling Addon = "tooling"
	Kured   Addon = "kured"

	Etcd    Component = "etcd"
	CoreDNS Component = "coredns"
	Pause   Component = "pause"
)

type ControlPlaneComponentsVersion struct {
	EtcdVersion    string
	CoreDNSVersion string
	PauseVersion   string
}

type ComponentsVersion struct {
	KubeletVersion string
}

type AddonsVersion struct {
	CiliumVersion  string
	ToolingVersion string
	KuredVersion   string
}

type KubernetesVersion struct {
	ControlPlaneComponentsVersion ControlPlaneComponentsVersion
	ComponentsVersion             ComponentsVersion
	AddonsVersion                 AddonsVersion
}

type KubernetesVersions map[string]KubernetesVersion

var (
	Versions = KubernetesVersions{
		// Old supported k8s versions
		// When you update the KubernetesVersion to a new one, add the current to the old supported versions
		"v1.14.1": KubernetesVersion{
			ControlPlaneComponentsVersion: ControlPlaneComponentsVersion{
				EtcdVersion:    "3.3.11",
				CoreDNSVersion: "1.2.6",
				PauseVersion:   "3.1",
			},
			ComponentsVersion: ComponentsVersion{
				KubeletVersion: "1.14.1",
			},
			AddonsVersion: AddonsVersion{
				CiliumVersion:  "1.5.3",
				ToolingVersion: "0.1.0",
				KuredVersion:   "1.2.0",
			},
		},
		// New (current) supported supported version
		fmt.Sprintf("v%s", CurrentVersion): KubernetesVersion{
			ControlPlaneComponentsVersion: ControlPlaneComponentsVersion{
				EtcdVersion:    currentETCDVersion,
				CoreDNSVersion: currentCoreDNSVersion,
				PauseVersion:   currentPauseVersion,
			},
			ComponentsVersion: ComponentsVersion{
				KubeletVersion: fmt.Sprintf("v%s", CurrentVersion),
			},
			AddonsVersion: AddonsVersion{
				CiliumVersion:  currentCiliumVersion,
				ToolingVersion: currentToolingVersion,
				KuredVersion:   currentKuredVersion,
			},
		},
	}
)

func CurrentComponentVersion(component Component) string {
	currentKubernetesVersion := Versions[fmt.Sprintf("v%s", CurrentVersion)]
	switch component {
	case Etcd:
		return currentKubernetesVersion.ControlPlaneComponentsVersion.EtcdVersion
	case CoreDNS:
		return currentKubernetesVersion.ControlPlaneComponentsVersion.CoreDNSVersion
	case Pause:
		return currentKubernetesVersion.ControlPlaneComponentsVersion.PauseVersion
	}
	log.Fatalf("unknown component %q", component)
	panic("unreachable")
}

func CurrentAddonVersion(addon Addon) string {
	currentKubernetesVersion := Versions[fmt.Sprintf("v%s", CurrentVersion)]
	switch addon {
	case Tooling:
		return currentKubernetesVersion.AddonsVersion.ToolingVersion
	case Cilium:
		return currentKubernetesVersion.AddonsVersion.CiliumVersion
	case Kured:
		return currentKubernetesVersion.AddonsVersion.KuredVersion
	}
	log.Fatalf("unknown addon %q", addon)
	panic("unreachable")
}

func availableVersionsForMap(versions KubernetesVersions) []*version.Version {
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
	return availableVersionsForMap(Versions)
}
