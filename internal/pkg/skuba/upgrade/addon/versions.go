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

package addon

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	skubaconfig "github.com/SUSE/skuba/internal/pkg/skuba/skuba"
)

type AddonVersionInfoUpdate struct {
	Current kubernetes.AddonsVersion
	Updated kubernetes.AddonsVersion
}

func UpdatedAddons(client clientset.Interface, clusterVersion *version.Version) (AddonVersionInfoUpdate, error) {
	skubaConfig, err := skubaconfig.GetSkubaConfiguration(client)
	if err != nil {
		return AddonVersionInfoUpdate{}, err
	}
	return UpdatedAddonsForAddonsVersion(clusterVersion, skubaConfig.AddonsVersion, kubernetes.AllAddonVersionsForClusterVersion), nil
}

func UpdatedAddonsForAddonsVersion(clusterVersion *version.Version, addonsVersion kubernetes.AddonsVersion, clusterAddonsKnownVersions kubernetes.ClusterAddonsKnownVersions) AddonVersionInfoUpdate {
	aviu := AddonVersionInfoUpdate{
		Current: kubernetes.AddonsVersion{},
		Updated: kubernetes.AddonsVersion{},
	}
	latestAddonVersions := clusterAddonsKnownVersions(clusterVersion)
	for addonName, addonLatestVersion := range latestAddonVersions {
		addonCurrentVersion := addonsVersion[addonName]
		aviu.Current[addonName] = addonCurrentVersion
		if addonCurrentVersion == nil || (addonLatestVersion.ManifestVersion > addonCurrentVersion.ManifestVersion) {
			aviu.Updated[addonName] = addonLatestVersion
		}
	}
	return aviu
}

func addonsByName(addons kubernetes.AddonsVersion) []kubernetes.Addon {
	sortedAddons := make([]kubernetes.Addon, len(addons))
	i := 0
	for addon := range addons {
		sortedAddons[i] = addon
		i++
	}
	sort.Slice(sortedAddons, func(i, j int) bool {
		return string(sortedAddons[i]) < string(sortedAddons[j])
	})
	return sortedAddons
}

// hasAddonVersionBump returns true if the version of a given addon is different in the current and
// the addon upgrade sets.
func hasAddonVersionBump(addonVersionInfoUpdate AddonVersionInfoUpdate, addon kubernetes.Addon) bool {
	return addonVersionInfoUpdate.Current[addon].Version != addonVersionInfoUpdate.Updated[addon].Version
}

func PrintAddonUpdates(updatedAddons AddonVersionInfoUpdate) {
	for _, addon := range addonsByName(updatedAddons.Updated) {
		if updatedAddons.Current[addon] == nil && updatedAddons.Updated[addon] != nil {
			if len(updatedAddons.Updated[addon].Version) > 0 {
				fmt.Printf("  - %s: %s (new addon)\n", addon, updatedAddons.Updated[addon].Version)
			} else {
				fmt.Printf("  - %s (new addon)\n", addon)
			}
			continue
		}

		if hasAddonVersionBump(updatedAddons, addon) {
			fmt.Printf("  - %s: %s -> %s\n", addon, updatedAddons.Current[addon].Version, updatedAddons.Updated[addon].Version)
		} else {
			if len(updatedAddons.Current[addon].Version) > 0 {
				fmt.Printf("  - %s: %s (manifest version from %d to %d)\n", addon,
					updatedAddons.Current[addon].Version,
					updatedAddons.Current[addon].ManifestVersion, updatedAddons.Updated[addon].ManifestVersion)
			} else {
				fmt.Printf("  - %s (manifest version from %d to %d)\n", addon,
					updatedAddons.Current[addon].ManifestVersion, updatedAddons.Updated[addon].ManifestVersion)
			}
		}
	}
}

func HasAddonUpdate(aviu AddonVersionInfoUpdate) bool {
	return len(aviu.Updated) > 0
}
