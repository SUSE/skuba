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

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
)

// Print out list of images that will be pulled
// This can be used as input to skopeo for mirroring in air-gapped scenarios
func Images() error {
	components := map[kubernetes.Component]string{
		kubernetes.Hyperkube: "hyperkube",
		kubernetes.Etcd:      "etcd",
		kubernetes.CoreDNS:   "coredns",
		kubernetes.Pause:     "pause",
		kubernetes.Tooling:   "tooling",
	}

	fmt.Printf("VERSION    IMAGE\n")
	for _, cv := range kubernetes.AvailableVersions() {
		for component, componentName := range components {
			fmt.Printf("%-10v %v\n", cv, images.GetGenericImage(skuba.ImageRepository, componentName,
				kubernetes.ComponentVersionForClusterVersion(component, cv)))
		}
		for addonName, addonVersion := range kubernetes.Versions[cv.String()].AddonsVersion {
			fmt.Printf("%-10v %v\n", cv, images.GetGenericImage(skuba.ImageRepository, string(addonName),
				addonVersion.Version))
		}
	}
	return nil
}
