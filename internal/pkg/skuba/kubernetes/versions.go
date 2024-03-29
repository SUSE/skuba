/*
 * Copyright (c) 2019,2020 SUSE LLC.
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
	"sort"

	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"

	"github.com/SUSE/skuba/pkg/skuba"
)

type Addon string
type Component string

const (
	Cilium        Addon = "cilium"
	Kured         Addon = "kured"
	Dex           Addon = "dex"
	Gangway       Addon = "gangway"
	MetricsServer Addon = "metrics-server"
	Kucero        Addon = "kucero"
	PSP           Addon = "psp"

	Kubelet          Component = "kubelet"
	ContainerRuntime Component = "cri-o"

	APIServer         Component = "apiserver"
	ControllerManager Component = "controllermanager"
	Scheduler         Component = "scheduler"
	Proxy             Component = "proxy"
	Hyperkube         Component = "hyperkube"
	Etcd              Component = "etcd"
	CoreDNS           Component = "coredns"
	Pause             Component = "pause"

	Tooling Component = "tooling"
)

type ComponentHostVersion struct {
	KubeletVersion          string
	ContainerRuntimeVersion string
}

type ContainerImageTag struct {
	Name string
	Tag  string
}

type ComponentContainerVersion map[Component]*ContainerImageTag

type AddonVersion struct {
	Version         string
	ManifestVersion uint
}

type AddonsVersion map[Addon]*AddonVersion

type KubernetesVersion struct {
	ComponentHostVersion      ComponentHostVersion
	ComponentContainerVersion ComponentContainerVersion
	AddonsVersion             AddonsVersion
}

type KubernetesVersions map[string]KubernetesVersion

type ClusterAddonsKnownVersions = func(clusterVersion *version.Version) AddonsVersion

var (
	supportedVersions = KubernetesVersions{
		"1.18.20": KubernetesVersion{
			ComponentHostVersion: ComponentHostVersion{
				KubeletVersion:          "1.18.20",
				ContainerRuntimeVersion: "1.18.4",
			},
			ComponentContainerVersion: ComponentContainerVersion{
				APIServer:         &ContainerImageTag{Name: "kube-apiserver", Tag: "v1.18.20"},
				ControllerManager: &ContainerImageTag{Name: "kube-controller-manager", Tag: "v1.18.20"},
				Scheduler:         &ContainerImageTag{Name: "kube-scheduler", Tag: "v1.18.20"},
				Proxy:             &ContainerImageTag{Name: "kube-proxy", Tag: "v1.18.20"},
				Etcd:              &ContainerImageTag{Name: "etcd", Tag: "3.4.13"},
				CoreDNS:           &ContainerImageTag{Name: "coredns", Tag: "1.6.7"},
				Pause:             &ContainerImageTag{Name: "pause", Tag: "3.2"},
				Tooling:           &ContainerImageTag{Name: "skuba-tooling", Tag: "0.1.0"},
			},
			AddonsVersion: AddonsVersion{
				Cilium:        &AddonVersion{"1.7.6-rev6", 4540},
				Kured:         &AddonVersion{"1.4.3-rev6", 4540},
				Dex:           &AddonVersion{"2.23.0-rev3", 4540},
				Gangway:       &AddonVersion{"3.1.0-rev7", 4540},
				MetricsServer: &AddonVersion{"0.3.6-rev3", 4540},
				Kucero:        &AddonVersion{"1.3.0-rev6", 4540},
				PSP:           &AddonVersion{"", 4540},
			},
		},
		"1.18.10": KubernetesVersion{
			ComponentHostVersion: ComponentHostVersion{
				KubeletVersion:          "1.18.10",
				ContainerRuntimeVersion: "1.18.4",
			},
			ComponentContainerVersion: ComponentContainerVersion{
				APIServer:         &ContainerImageTag{Name: "kube-apiserver", Tag: "v1.18.10"},
				ControllerManager: &ContainerImageTag{Name: "kube-controller-manager", Tag: "v1.18.10"},
				Scheduler:         &ContainerImageTag{Name: "kube-scheduler", Tag: "v1.18.10"},
				Proxy:             &ContainerImageTag{Name: "kube-proxy", Tag: "v1.18.10"},
				Etcd:              &ContainerImageTag{Name: "etcd", Tag: "3.4.13"},
				CoreDNS:           &ContainerImageTag{Name: "coredns", Tag: "1.6.7"},
				Pause:             &ContainerImageTag{Name: "pause", Tag: "3.2"},
				Tooling:           &ContainerImageTag{Name: "skuba-tooling", Tag: "0.1.0"},
			},
			AddonsVersion: AddonsVersion{
				Cilium:        &AddonVersion{"1.7.6-rev6", 4540},
				Kured:         &AddonVersion{"1.4.3-rev6", 4540},
				Dex:           &AddonVersion{"2.23.0-rev3", 4540},
				Gangway:       &AddonVersion{"3.1.0-rev7", 4540},
				MetricsServer: &AddonVersion{"0.3.6-rev3", 4540},
				Kucero:        &AddonVersion{"1.3.0-rev6", 4540},
				PSP:           &AddonVersion{"", 4540},
			},
		},
		"1.18.6": KubernetesVersion{
			ComponentHostVersion: ComponentHostVersion{
				KubeletVersion:          "1.18.6",
				ContainerRuntimeVersion: "1.18.2",
			},
			ComponentContainerVersion: ComponentContainerVersion{
				APIServer:         &ContainerImageTag{Name: "kube-apiserver", Tag: "v1.18.6"},
				ControllerManager: &ContainerImageTag{Name: "kube-controller-manager", Tag: "v1.18.6"},
				Scheduler:         &ContainerImageTag{Name: "kube-scheduler", Tag: "v1.18.6"},
				Proxy:             &ContainerImageTag{Name: "kube-proxy", Tag: "v1.18.6"},
				Etcd:              &ContainerImageTag{Name: "etcd", Tag: "3.4.3"},
				CoreDNS:           &ContainerImageTag{Name: "coredns", Tag: "1.6.7"},
				Pause:             &ContainerImageTag{Name: "pause", Tag: "3.2"},
				Tooling:           &ContainerImageTag{Name: "skuba-tooling", Tag: "0.1.0"},
			},
			AddonsVersion: AddonsVersion{
				Cilium:        &AddonVersion{"1.7.6-rev3", 4501},
				Kured:         &AddonVersion{"1.4.3", 4500},
				Dex:           &AddonVersion{"2.23.0", 4500},
				Gangway:       &AddonVersion{"3.1.0-rev5", 4500},
				MetricsServer: &AddonVersion{"0.3.6", 4500},
				Kucero:        &AddonVersion{"1.1.1", 4500},
				PSP:           &AddonVersion{"", 4500},
			},
		},
		"1.17.13": KubernetesVersion{
			ComponentHostVersion: ComponentHostVersion{
				KubeletVersion:          "1.17.13",
				ContainerRuntimeVersion: "1.16.1",
			},
			ComponentContainerVersion: ComponentContainerVersion{
				APIServer:         &ContainerImageTag{Name: "hyperkube", Tag: "v1.17.13"},
				ControllerManager: &ContainerImageTag{Name: "hyperkube", Tag: "v1.17.13"},
				Scheduler:         &ContainerImageTag{Name: "hyperkube", Tag: "v1.17.13"},
				Proxy:             &ContainerImageTag{Name: "hyperkube", Tag: "v1.17.13"},
				Etcd:              &ContainerImageTag{Name: "etcd", Tag: "3.4.13"},
				CoreDNS:           &ContainerImageTag{Name: "coredns", Tag: "1.6.7"},
				Pause:             &ContainerImageTag{Name: "pause", Tag: "3.1"},
				Tooling:           &ContainerImageTag{Name: "skuba-tooling", Tag: "0.1.0"},
			},
			AddonsVersion: AddonsVersion{
				Cilium:        &AddonVersion{"1.6.6-rev5", 4},
				Kured:         &AddonVersion{"1.3.0", 4},
				Dex:           &AddonVersion{"2.16.0-rev6", 7},
				Gangway:       &AddonVersion{"3.1.0-rev4", 6},
				MetricsServer: &AddonVersion{"0.3.6", 1},
				Kucero:        &AddonVersion{"1.3.0", 0},
				PSP:           &AddonVersion{"", 4},
			},
		},
		"1.17.4": KubernetesVersion{
			ComponentHostVersion: ComponentHostVersion{
				KubeletVersion:          "1.17.4",
				ContainerRuntimeVersion: "1.16.1",
			},
			ComponentContainerVersion: ComponentContainerVersion{
				APIServer:         &ContainerImageTag{Name: "hyperkube", Tag: "v1.17.4"},
				ControllerManager: &ContainerImageTag{Name: "hyperkube", Tag: "v1.17.4"},
				Scheduler:         &ContainerImageTag{Name: "hyperkube", Tag: "v1.17.4"},
				Proxy:             &ContainerImageTag{Name: "hyperkube", Tag: "v1.17.4"},
				Etcd:              &ContainerImageTag{Name: "etcd", Tag: "3.4.3"},
				CoreDNS:           &ContainerImageTag{Name: "coredns", Tag: "1.6.5"},
				Pause:             &ContainerImageTag{Name: "pause", Tag: "3.1"},
				Tooling:           &ContainerImageTag{Name: "skuba-tooling", Tag: "0.1.0"},
			},
			AddonsVersion: AddonsVersion{
				Cilium:        &AddonVersion{"1.6.6-rev5", 4},
				Kured:         &AddonVersion{"1.3.0", 4},
				Dex:           &AddonVersion{"2.16.0-rev6", 7},
				Gangway:       &AddonVersion{"3.1.0-rev4", 6},
				MetricsServer: &AddonVersion{"0.3.6", 1},
				PSP:           &AddonVersion{"", 4},
			},
		},
		"1.16.2": KubernetesVersion{
			ComponentHostVersion: ComponentHostVersion{
				KubeletVersion:          "1.16.2",
				ContainerRuntimeVersion: "1.16.1",
			},
			ComponentContainerVersion: ComponentContainerVersion{
				APIServer:         &ContainerImageTag{Name: "hyperkube", Tag: "v1.16.2"},
				ControllerManager: &ContainerImageTag{Name: "hyperkube", Tag: "v1.16.2"},
				Scheduler:         &ContainerImageTag{Name: "hyperkube", Tag: "v1.16.2"},
				Proxy:             &ContainerImageTag{Name: "hyperkube", Tag: "v1.16.2"},
				Etcd:              &ContainerImageTag{Name: "etcd", Tag: "3.3.15"},
				CoreDNS:           &ContainerImageTag{Name: "coredns", Tag: "1.6.2"},
				Pause:             &ContainerImageTag{Name: "pause", Tag: "3.1"},
				Tooling:           &ContainerImageTag{Name: "skuba-tooling", Tag: "0.1.0"},
			},
			AddonsVersion: AddonsVersion{
				Cilium:        &AddonVersion{"1.5.3", 3},
				Kured:         &AddonVersion{"1.3.0", 4},
				Dex:           &AddonVersion{"2.16.0", 6},
				Gangway:       &AddonVersion{"3.1.0-rev4", 5},
				MetricsServer: &AddonVersion{"0.3.6", 1},
				PSP:           &AddonVersion{"", 2},
			},
		},
		"1.15.2": KubernetesVersion{
			ComponentHostVersion: ComponentHostVersion{
				KubeletVersion:          "1.15.2",
				ContainerRuntimeVersion: "1.15.2",
			},
			ComponentContainerVersion: ComponentContainerVersion{
				APIServer:         &ContainerImageTag{Name: "hyperkube", Tag: "v1.15.2"},
				ControllerManager: &ContainerImageTag{Name: "hyperkube", Tag: "v1.15.2"},
				Scheduler:         &ContainerImageTag{Name: "hyperkube", Tag: "v1.15.2"},
				Proxy:             &ContainerImageTag{Name: "hyperkube", Tag: "v1.15.2"},
				Etcd:              &ContainerImageTag{Name: "etcd", Tag: "3.3.11"},
				CoreDNS:           &ContainerImageTag{Name: "coredns", Tag: "1.3.1"},
				Pause:             &ContainerImageTag{Name: "pause", Tag: "3.1"},
				Tooling:           &ContainerImageTag{Name: "skuba-tooling", Tag: "0.1.0"},
			},
			AddonsVersion: AddonsVersion{
				Cilium:  &AddonVersion{"1.5.3", 3},
				Kured:   &AddonVersion{"1.2.0-rev4", 2},
				Dex:     &AddonVersion{"2.16.0", 6},
				Gangway: &AddonVersion{"3.1.0-rev4", 5},
				PSP:     &AddonVersion{"", 1},
			},
		},
	}
)

func ComponentVersionWithAvailableVersions(component Component, clusterVersion *version.Version, availableVersions KubernetesVersions) string {
	currentKubernetesVersion := availableVersions[clusterVersion.String()]
	switch component {
	case Kubelet:
		return currentKubernetesVersion.ComponentHostVersion.KubeletVersion
	case ContainerRuntime:
		return currentKubernetesVersion.ComponentHostVersion.ContainerRuntimeVersion
	default:
		if componentVersion, found := currentKubernetesVersion.ComponentContainerVersion[component]; found {
			return componentVersion.Tag
		}
	}
	klog.Errorf("unknown component %q version", component)
	return ""
}

func ComponentVersionForClusterVersion(component Component, clusterVersion *version.Version) string {
	return ComponentVersionWithAvailableVersions(component, clusterVersion, supportedVersions)
}

func AllComponentContainerImagesForClusterVersion(clusterVersion *version.Version) []Component {
	currentKubernetesVersion := supportedVersions[clusterVersion.String()]

	components := make([]Component, 0)
	for component := range currentKubernetesVersion.ComponentContainerVersion {
		components = append(components, component)
	}
	return components
}

func ComponentContainerImageForClusterVersion(component Component, clusterVersion *version.Version) string {
	currentKubernetesVersion := supportedVersions[clusterVersion.String()]
	if componentDetails, found := currentKubernetesVersion.ComponentContainerVersion[component]; found {
		return images.GetGenericImage(skuba.ImageRepository(clusterVersion), componentDetails.Name, componentDetails.Tag)
	}
	klog.Errorf("unknown component %q container image", component)
	return ""
}

func AddonVersionForClusterVersion(addon Addon, clusterVersion *version.Version) *AddonVersion {
	currentKubernetesVersion := supportedVersions[clusterVersion.String()]
	if addonVersion, found := currentKubernetesVersion.AddonsVersion[addon]; found {
		return addonVersion
	}
	return nil
}

func AllAddonVersionsForClusterVersion(clusterVersion *version.Version) AddonsVersion {
	return supportedVersions[clusterVersion.String()].AddonsVersion
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
	return AvailableVersionsForMap(supportedVersions)
}

// LatestVersion return the latest Kubernetes supported version
func LatestVersion() *version.Version {
	availableVersions := AvailableVersions()
	return availableVersions[len(availableVersions)-1]
}

// IsVersionAvailable returns if a specific kubernetes version is available
func IsVersionAvailable(kubernetesVersion *version.Version) bool {
	_, ok := supportedVersions[kubernetesVersion.String()]
	return ok
}

// MajorMinorVersion returns a KubernetesVersion without the patch level
func MajorMinorVersion(kubernetesVersion *version.Version) string {
	return fmt.Sprintf("%d.%d", kubernetesVersion.Major(), kubernetesVersion.Minor())
}
