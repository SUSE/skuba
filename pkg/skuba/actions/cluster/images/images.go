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
// Ensure images only appear once
func Images() error {
	fmt.Printf("VERSION    IMAGE\n")

	for _, version := range kubernetes.AvailableVersions() {
		imagesEncountered := map[string]bool{}
		for _, component := range kubernetes.AllComponentContainerImagesForClusterVersion(version) {
			imagesEncountered[kubernetes.ComponentContainerImageForClusterVersion(component, version)] = true
		}

		for addonName, addon := range addons.Addons {
			addonVersion := kubernetes.AddonVersionForClusterVersion(addonName, version)
			if addonVersion == nil {
				continue
			}
			for _, addonImageLoc := range addon.Images(addonVersion.Version) {
				imagesEncountered[addonImageLoc] = true
			}
		}

		for imagelocation := range imagesEncountered {
			fmt.Printf("%-10v %v\n", version, imagelocation)
		}
	}
	return nil
}
