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
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	skubaconfig "github.com/SUSE/skuba/internal/pkg/skuba/skuba"
)

type AddonVersionInfoUpdate struct {
	Current kubernetes.AddonsVersion
	Updated kubernetes.AddonsVersion
}

func UpdatedAddons() (AddonVersionInfoUpdate, error) {
	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return AddonVersionInfoUpdate{}, err
	}
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return AddonVersionInfoUpdate{}, err
	}
	currentVersion := currentClusterVersion.String()
	latestAddonVersions := kubernetes.Versions[currentVersion].AddonsVersion

	skubaConfig, err := skubaconfig.GetSkubaConfiguration()
	if err != nil {
		return AddonVersionInfoUpdate{}, err
	}
	aviu := AddonVersionInfoUpdate{
		Current: kubernetes.AddonsVersion{},
		Updated: kubernetes.AddonsVersion{},
	}

	for addonName, version := range latestAddonVersions {
		skubaConfigVersion := skubaConfig.AddonsVersion[addonName]
		aviu.Current[addonName] = skubaConfigVersion
		if version.Version > skubaConfigVersion.Version || version.ManifestVersion > skubaConfigVersion.ManifestVersion {
			aviu.Updated[addonName] = version
		}
	}
	return aviu, nil
}

func HasAddonUpdate(aviu AddonVersionInfoUpdate) bool {
	for addon, _ := range aviu.Updated {
		if HasAddonManifestUpdateWithAddon(aviu, addon) || HasAddonVersionUpdateWithAddon(aviu, addon) {
			return true
		}
	}
	return false
}

func HasAddonManifestUpdateWithAddon(aviu AddonVersionInfoUpdate, addon kubernetes.Addon) bool {
	return aviu.Updated[addon].ManifestVersion > aviu.Current[addon].ManifestVersion
}

func HasAddonVersionUpdateWithAddon(aviu AddonVersionInfoUpdate, addon kubernetes.Addon) bool {
	return aviu.Updated[addon].Version > aviu.Current[addon].Version
}
