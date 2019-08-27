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

package cluster

import (
	"fmt"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

// Print out list of images that will be pulled
// This can be used as input to skopeo for mirroring in air-gapped scenarios
func Images() error {
	for _, ver := range kubernetes.AvailableVersions() {
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "pause",
			kubernetes.ComponentVersionForClusterVersion(kubernetes.Pause,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "hyperkube",
			kubernetes.ComponentVersionForClusterVersion(kubernetes.Hyperkube,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "etcd",
			kubernetes.ComponentVersionForClusterVersion(kubernetes.Etcd,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "coredns",
			kubernetes.ComponentVersionForClusterVersion(kubernetes.CoreDNS,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "kured",
			kubernetes.AddonVersionForClusterVersion(kubernetes.Kured,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "cilium",
			kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "cilium-init",
			kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "cilium-operator",
			kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "skuba-tooling",
			kubernetes.AddonVersionForClusterVersion(kubernetes.Tooling,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "caasp-dex",
			kubernetes.AddonVersionForClusterVersion(kubernetes.Dex,ver)))
		fmt.Println(ver,images.GetGenericImage(skuba.ImageRepository, "gangway",
			kubernetes.AddonVersionForClusterVersion(kubernetes.Gangway,ver)))
	}

	return nil
}
