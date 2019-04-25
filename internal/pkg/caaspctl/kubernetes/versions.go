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

type Component string

const (
	Etcd    Component = "etcd"
	CoreDNS Component = "coredns"
	Pause   Component = "pause"
	Cilium  Component = "cilium"
)

type ControlPlaneComponentsVersion struct {
	EtcdVersion    string
	CoreDNSVersion string
	PauseVersion   string
	CiliumVersion  string
}

type ComponentsVersion struct {
	KubeletVersion string
}

type KubernetesVersion struct {
	ControlPlaneComponentsVersion ControlPlaneComponentsVersion
	ComponentsVersion             ComponentsVersion
}

const (
	CurrentVersion = "v1.14.0"
)

var (
	Versions = map[string]KubernetesVersion{
		"v1.14.0": KubernetesVersion{
			ControlPlaneComponentsVersion: ControlPlaneComponentsVersion{
				EtcdVersion:    "3.3.1",
				CoreDNSVersion: "1.2.6",
				PauseVersion:   "3.1",
				CiliumVersion:  "1.4.2",
			},
			ComponentsVersion: ComponentsVersion{
				KubeletVersion: "1.14.0",
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
	case Cilium:
		return currentKubernetesVersion.ControlPlaneComponentsVersion.CiliumVersion
	}
	log.Fatalf("unknown component %q", component)
	panic("unreachable")
}
