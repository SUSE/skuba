/*
 * Copyright (c) 2019 SUSE LLC. All rights reserved.
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

const (
	CurrentVersion = "v1.14.1"
)

var (
	Versions = map[string]KubernetesVersion{
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
				CiliumVersion:  "1.4.2",
				ToolingVersion: "0.1.0",
				KuredVersion:   "1.2.0",
			},
		},
	}
)

func CurrentComponentVersion(component Component) string {
	currentKubernetesVersion := Versions[CurrentVersion]
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
	currentKubernetesVersion := Versions[CurrentVersion]
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
