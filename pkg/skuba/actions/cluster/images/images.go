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

	"github.com/SUSE/skuba/internal/pkg/skuba/addons"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

// Print out a list of images that will use
// This can be used as input to skopeo for mirroring in air-gapped scenarios
func Images() error {
	fmt.Printf("VERSION    IMAGE\n")
	for _, version := range kubernetes.AvailableVersions() {
		for _, component := range kubernetes.AllComponentContainerImagesForClusterVersion(version) {
			fmt.Printf("%-10v %v\n", version, kubernetes.ComponentContainerImageForClusterVersion(component, version))
		}

		for addonName, addon := range addons.Addons {
			addonVersion := kubernetes.AddonVersionForClusterVersion(addonName, version)
			if addonVersion == nil {
				continue
			}
			imageList := addon.Images(addonVersion.Version)
			for _, image := range imageList {
				fmt.Printf("%-10v %v\n", version, image)
			}
		}
	}
	return nil
}
