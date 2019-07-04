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
	"log"
	"sort"

	"k8s.io/apimachinery/pkg/util/version"
)

type Addon string
type Component string

const (
	Cilium  Addon = "cilium"
	Tooling Addon = "tooling"
	Kured   Addon = "kured"
	Gangway Addon = "gangway"

	Hyperkube Component = "hyperkube"
	Etcd      Component = "etcd"
	CoreDNS   Component = "coredns"
	Pause     Component = "pause"
)

type ControlPlaneComponentsVersion struct {
	HyperkubeVersion string
	EtcdVersion      string
	CoreDNSVersion   string
	PauseVersion     string
}

type ComponentsVersion struct {
	KubeletVersion string
}

type AddonsVersion struct {
	CiliumVersion  string
	ToolingVersion string
	KuredVersion   string
	GangwayVersion string
}

type KubernetesVersion struct {
	ControlPlaneComponentsVersion ControlPlaneComponentsVersion
	ComponentsVersion             ComponentsVersion
	AddonsVersion                 AddonsVersion
}

type KubernetesVersions map[string]KubernetesVersion

var (
	Versions = KubernetesVersions{
		"1.14.1": KubernetesVersion{
			ControlPlaneComponentsVersion: ControlPlaneComponentsVersion{
				HyperkubeVersion: "v1.14.1",
				EtcdVersion:      "3.3.11",
				CoreDNSVersion:   "1.2.6",
				PauseVersion:     "3.1",
			},
			ComponentsVersion: ComponentsVersion{
				KubeletVersion: "1.14.1",
			},
			AddonsVersion: AddonsVersion{
				CiliumVersion:  "1.5.3",
				ToolingVersion: "0.1.0",
				KuredVersion:   "1.2.0",
				GangwayVersion: "3.1.0",
			},
		},
	}
)

func CurrentComponentVersion(component Component) string {
	currentKubernetesVersion := Versions[LatestVersion().String()]
	switch component {
	case Hyperkube:
		return currentKubernetesVersion.ControlPlaneComponentsVersion.HyperkubeVersion
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
	currentKubernetesVersion := Versions[LatestVersion().String()]
	switch addon {
	case Tooling:
		return currentKubernetesVersion.AddonsVersion.ToolingVersion
	case Cilium:
		return currentKubernetesVersion.AddonsVersion.CiliumVersion
	case Kured:
		return currentKubernetesVersion.AddonsVersion.KuredVersion
	case Gangway:
		return currentKubernetesVersion.AddonsVersion.GangwayVersion
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

// LatestVersion return the latest Kubernetes supported version
func LatestVersion() *version.Version {
	availableVersions := AvailableVersions()
	return availableVersions[len(availableVersions)-1]
}
