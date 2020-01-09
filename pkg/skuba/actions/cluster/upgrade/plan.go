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

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/upgrade/addon"
	upgradecluster "github.com/SUSE/skuba/internal/pkg/skuba/upgrade/cluster"
)

func Plan(client clientset.Interface) error {
	return plan(client, kubernetes.AvailableVersions())
}

func plan(client clientset.Interface, availableVersions []*version.Version) error {
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return err
	}
	currentVersion := currentClusterVersion.String()
	latestVersion := availableVersions[len(availableVersions)-1].String()
	fmt.Printf("Current Kubernetes cluster version: %s\n", currentVersion)
	fmt.Printf("Latest Kubernetes version: %s\n", latestVersion)
	fmt.Println()

	if err := planPrePlatformUpgrade(client, currentClusterVersion); err != nil {
		return err
	}
	upgradePath, err := planPlatformUpgrade(client, currentClusterVersion, availableVersions)
	if err != nil {
		return err
	}
	if len(upgradePath) > 0 {
		if err := planPostPlatformUpgrade(currentClusterVersion, upgradePath); err != nil {
			return err
		}
	}

	return nil
}

func planPrePlatformUpgrade(client clientset.Interface, currentClusterVersion *version.Version) error {
	if err := checkDriftedNodes(client); err != nil {
		return err
	}
	return checkUpdatedAddons(client, currentClusterVersion)
}

func checkDriftedNodes(client clientset.Interface) error {
	driftedNodes, err := upgradecluster.DriftedNodes(client)
	if err != nil {
		return err
	}
	if len(driftedNodes) > 0 {
		fmt.Println()
		fmt.Println("WARNING: Incomplete upgrade detected:")
		for _, node := range driftedNodes {
			fmt.Printf("  - %s is running kubelet %s and is not cordoned\n", node.Node.ObjectMeta.Name, node.KubeletVersion.String())
		}
	}
	return nil
}

// checkUpdatedAddons compares the list of the current addons in the cluster with the latest
// available versions of addons for the current cluster version.
func checkUpdatedAddons(client clientset.Interface, currentClusterVersion *version.Version) error {
	updatedAddons, err := addon.UpdatedAddons(client, currentClusterVersion)
	if err != nil {
		return err
	}
	fmt.Println()
	if addon.HasAddonUpdate(updatedAddons) {
		fmt.Printf("Addon upgrades for %s:\n", currentClusterVersion)
		addon.PrintAddonUpdates(updatedAddons)
	} else {
		fmt.Printf("Addons at the current cluster version %s are already up to date.\n", currentClusterVersion.String())
		fmt.Println()
		fmt.Println("There is no need to run `skuba addon upgrade apply` before starting the platform upgrade.")
	}
	return nil
}

func planPlatformUpgrade(client clientset.Interface, currentClusterVersion *version.Version, availableVersions []*version.Version) ([]*version.Version, error) {
	upgradePath := []*version.Version{}
	versionCompare, err := currentClusterVersion.Compare(kubernetes.LatestVersion().String())
	if err != nil {
		return upgradePath, err
	}
	if versionCompare != 0 {
		// Platform is not up to date, print upgrade path
		upgradePath, err := upgradecluster.UpgradePathWithAvailableVersions(currentClusterVersion, availableVersions)
		if err != nil {
			return upgradePath, err
		}
		if len(upgradePath) == 0 {
			return upgradePath, errors.Errorf("cannot infer how to upgrade from %s to %s", currentClusterVersion, kubernetes.LatestVersion())
		}
		fmt.Printf("Upgrade path to update from %s to %s:\n", currentClusterVersion, kubernetes.LatestVersion())
		tmpVersion := currentClusterVersion.String()
		for _, version := range upgradePath {
			fmt.Printf(" - %s -> %s\n", tmpVersion, version.String())
			tmpVersion = version.String()
		}
	} else {
		// Platform is up to date, print node versioning information
		nodeVersionInfoMap, err := kubernetes.AllNodesVersioningInfo(client)
		if err != nil {
			return upgradePath, err
		}
		fmt.Println()
		fmt.Printf("Upgrade node status to cluster version %s:\n", currentClusterVersion.String())
		for nodeName, nodeVersionInfo := range nodeVersionInfoMap {
			if nodeVersionInfo.EqualsClusterVersion(currentClusterVersion) {
				fmt.Printf("- %s: up to date", nodeName)
			} else {
				fmt.Printf("- %s: current version: %s (needs upgrade)", nodeName, nodeVersionInfo.String())
			}
			if nodeVersionInfo.Unschedulable() {
				fmt.Println("; unschedulable, ignored")
			} else {
				fmt.Println()
			}
		}
	}
	return upgradePath, nil
}

func planPostPlatformUpgrade(currentClusterVersion *version.Version, upgradePath []*version.Version) error {
	nextClusterVersion := upgradePath[0]
	if err := checkUpdatedAddonsFromClusterVersion(currentClusterVersion, nextClusterVersion); err != nil {
		return err
	}
	return nil
}

// checkUpdatedAddonsFromClusterVersion compares the latest available versions of addons for
// currentClusterVersion with the list of the latest available versions of addons for
// nextClusterVersion. It does not check the current addons versions in the cluster.
func checkUpdatedAddonsFromClusterVersion(currentClusterVersion *version.Version, nextClusterVersion *version.Version) error {
	// Assuming we are at the latest addon versions of the current cluster version
	latestAddonsForCurrentClusterVersion := kubernetes.AllAddonVersionsForClusterVersion(currentClusterVersion)
	updatedAddons := addon.UpdatedAddonsForAddonsVersion(nextClusterVersion, latestAddonsForCurrentClusterVersion)
	fmt.Println()
	if addon.HasAddonUpdate(updatedAddons) {
		fmt.Printf("Addon upgrades from %s to %s:\n", currentClusterVersion, nextClusterVersion.String())
		addon.PrintAddonUpdates(updatedAddons)
	} else {
		fmt.Printf("Addons for next cluster version %s are already up to date.\n", nextClusterVersion.String())
		fmt.Println()
		fmt.Println("There is no need to run `skuba addon upgrade apply` after you have completed the platform upgrade.")
	}
	return nil
}
