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

package upgrade

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	skubaconfig "github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	"github.com/SUSE/skuba/internal/pkg/skuba/upgrade/addon"
	upgradecluster "github.com/SUSE/skuba/internal/pkg/skuba/upgrade/cluster"
)

func Plan(client clientset.Interface) error {
	return plan(client, kubernetes.AvailableVersions(), kubernetes.AllAddonVersionsForClusterVersion)
}

func plan(client clientset.Interface, availableVersions []*version.Version, clusterAddonsKnownVersions kubernetes.ClusterAddonsKnownVersions) error {
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return err
	}
	currentVersion := currentClusterVersion.String()
	latestClusterVersion := availableVersions[len(availableVersions)-1]
	fmt.Printf("Current Kubernetes cluster version: %s\n", currentVersion)
	fmt.Printf("Latest Kubernetes version: %s\n", latestClusterVersion.String())

	upgradePath, err := calculateUpgradePath(currentClusterVersion, availableVersions)
	if err != nil {
		return err
	}

	if err := checkPlatformUpgradeFeasible(currentClusterVersion, latestClusterVersion, upgradePath); err != nil {
		return err
	}

	currentAddonVersionInfoUpdate, err := checkUpdatedAddons(client, currentClusterVersion, clusterAddonsKnownVersions)
	if err != nil {
		return err
	}

	var nextClusterVersion *version.Version
	if len(upgradePath) > 0 {
		nextClusterVersion = upgradePath[0]
	}

	nodeVersionInfoMap, err := kubernetes.AllNodesVersioningInfo(client)
	if err != nil {
		return err
	}

	allNodesMatchClusterVersion := kubernetes.AllNodesMatchClusterVersionWithVersioningInfo(nodeVersionInfoMap, currentClusterVersion)

	fmt.Println()
	if !allNodesMatchClusterVersion {
		fmt.Printf("Some nodes do not match the current cluster version (%s):\n", currentClusterVersion.String())

		sortedNodeNames := []string{}
		for nodeName := range nodeVersionInfoMap {
			sortedNodeNames = append(sortedNodeNames, nodeName)
		}
		sort.Strings(sortedNodeNames)
		for _, nodeName := range sortedNodeNames {
			nodeVersionInfo := nodeVersionInfoMap[nodeName]
			if nodeVersionInfo.EqualsClusterVersion(currentClusterVersion) {
				fmt.Printf("  - %s: up to date", nodeName)
			} else if !nodeVersionInfo.ToleratesClusterVersion(currentClusterVersion) {
				fmt.Printf("  - %s; current version: %s (upgrade required)", nodeName, nodeVersionInfo.String())
			} else {
				fmt.Printf("  - %s; current version: %s (upgrade suggested)", nodeName, nodeVersionInfo.String())
			}
			if nodeVersionInfo.Unschedulable() {
				fmt.Println("; unschedulable, ignored")
			} else {
				fmt.Println()
			}
		}
	} else {
		fmt.Printf("All nodes match the current cluster version: %s.\n", currentClusterVersion.String())
	}

	planPrePlatformUpgrade(currentClusterVersion, nextClusterVersion, currentAddonVersionInfoUpdate)
	hasPlatformUpgrade := len(upgradePath) > 0
	if hasPlatformUpgrade {
		if err := planPlatformUpgrade(currentClusterVersion, latestClusterVersion, upgradePath); err != nil {
			return err
		}
		if err := planPostPlatformUpgrade(currentClusterVersion, nextClusterVersion, clusterAddonsKnownVersions); err != nil {
			return err
		}
	}

	return nil
}

func planPrePlatformUpgrade(currentClusterVersion *version.Version, nextClusterVersion *version.Version, addonVersionInfoUpdate addon.AddonVersionInfoUpdate) {
	fmt.Println()
	if addon.HasAddonUpdate(addonVersionInfoUpdate) {
		fmt.Printf("Addon upgrades for %s:\n", currentClusterVersion)
		addon.PrintAddonUpdates(addonVersionInfoUpdate)
		if nextClusterVersion != nil {
			fmt.Println()
			fmt.Println("It is required to run 'skuba addon upgrade apply' before starting the platform upgrade.")
		}
	} else {
		fmt.Printf("Addons at the current cluster version %s are up to date.\n", currentClusterVersion.String())
		if nextClusterVersion != nil {
			fmt.Println("There is no need to run 'skuba addon upgrade apply' before starting the platform upgrade.")
		}
	}
}

// checkUpdatedAddons compares the list of the current addons in the cluster with the latest
// available versions of addons for the current cluster version.
func checkUpdatedAddons(client clientset.Interface, currentClusterVersion *version.Version, clusterAddonsKnownVersions kubernetes.ClusterAddonsKnownVersions) (addon.AddonVersionInfoUpdate, error) {
	skubaConfig, err := skubaconfig.GetSkubaConfiguration(client)
	if err != nil {
		return addon.AddonVersionInfoUpdate{}, err
	}
	return addon.UpdatedAddonsForAddonsVersion(currentClusterVersion, skubaConfig.AddonsVersion, clusterAddonsKnownVersions), nil
}

func calculateUpgradePath(currentClusterVersion *version.Version, availableVersions []*version.Version) ([]*version.Version, error) {
	latestClusterVersion := availableVersions[len(availableVersions)-1]
	versionCompare, err := currentClusterVersion.Compare(latestClusterVersion.String())
	if err != nil {
		return []*version.Version{}, err
	}
	if versionCompare == 0 {
		return []*version.Version{}, nil
	}
	return upgradecluster.UpgradePathWithAvailableVersions(currentClusterVersion, availableVersions)
}

func checkPlatformUpgradeFeasible(currentClusterVersion *version.Version, latestClusterVersion *version.Version, upgradePath []*version.Version) error {
	versionCompare, err := currentClusterVersion.Compare(latestClusterVersion.String())
	if err != nil {
		return err
	}
	if versionCompare != 0 && len(upgradePath) == 0 {
		return errors.Errorf("cannot infer how to upgrade from %s to %s", currentClusterVersion, latestClusterVersion)
	}
	return nil
}

func planPlatformUpgrade(currentClusterVersion *version.Version, latestClusterVersion *version.Version, upgradePath []*version.Version) error {
	versionCompare, err := currentClusterVersion.Compare(latestClusterVersion.String())
	if err != nil {
		return err
	}
	if versionCompare != 0 {
		fmt.Println()
		fmt.Printf("Upgrade path to update from %s to %s:\n", currentClusterVersion, latestClusterVersion)
		tmpVersion := currentClusterVersion.String()
		for _, version := range upgradePath {
			fmt.Printf("  - %s -> %s\n", tmpVersion, version.String())
			tmpVersion = version.String()
		}
	}
	return nil
}

func planPostPlatformUpgrade(currentClusterVersion *version.Version, nextClusterVersion *version.Version, clusterAddonsKnownVersions kubernetes.ClusterAddonsKnownVersions) error {
	if err := checkUpdatedAddonsFromClusterVersion(currentClusterVersion, nextClusterVersion, clusterAddonsKnownVersions); err != nil {
		return err
	}
	return nil
}

// checkUpdatedAddonsFromClusterVersion compares the latest available versions of addons for
// currentClusterVersion with the list of the latest available versions of addons for
// nextClusterVersion. It does not check the current addons versions in the cluster.
func checkUpdatedAddonsFromClusterVersion(currentClusterVersion *version.Version, nextClusterVersion *version.Version, clusterAddonsKnownVersions kubernetes.ClusterAddonsKnownVersions) error {
	// Assuming we are at the latest addon versions of the current cluster version
	latestAddonsForCurrentClusterVersion := clusterAddonsKnownVersions(currentClusterVersion)
	updatedAddons := addon.UpdatedAddonsForAddonsVersion(nextClusterVersion, latestAddonsForCurrentClusterVersion, clusterAddonsKnownVersions)
	fmt.Println()
	if addon.HasAddonUpdate(updatedAddons) {
		fmt.Printf("Addon upgrades from %s to %s:\n", currentClusterVersion, nextClusterVersion.String())
		addon.PrintAddonUpdates(updatedAddons)
		if nextClusterVersion != nil {
			fmt.Println()
			fmt.Println("It is required to run 'skuba addon upgrade apply' after you have completed the platform upgrade.")
		}
	} else {
		if nextClusterVersion != nil {
			fmt.Println("There is no need to run 'skuba addon upgrade apply' after you have completed the platform upgrade.")
		}
	}
	return nil
}
