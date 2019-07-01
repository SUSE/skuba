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

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/upgrade"
	"github.com/SUSE/skuba/pkg/skuba"
)

func Plan() error {
	fmt.Printf("%s\n", skuba.CurrentVersion().String())

	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion()
	if err != nil {
		return err
	}
	currentVersion := currentClusterVersion.String()
	latestVersion := kubernetes.LatestVersion().String()
	fmt.Printf("Current Kubernetes cluster version: %s\n", currentVersion)
	fmt.Printf("Latest Kubernetes version: %s\n", latestVersion)

	if currentVersion == latestVersion {
		fmt.Println("Congratulations! You are already at the latest version available")
		return nil
	}

	upgradePath, err := upgrade.UpgradePath()
	if err != nil {
		return err
	}

	if len(upgradePath) == 0 {
		return errors.Errorf("cannot infer how to upgrade from %s to %s", currentVersion, latestVersion)
	}

	fmt.Printf("Upgrade path to update from %s to %s:\n", currentVersion, latestVersion)
	tmpVersion := currentVersion
	for _, version := range upgradePath {
		fmt.Printf(" - %s -> %s\n", tmpVersion, version.String())
		tmpVersion = version.String()
	}

	driftedNodes, err := upgrade.DriftedNodes()
	if err != nil {
		return err
	}
	if len(driftedNodes) > 0 {
		fmt.Println()
		fmt.Println("WARNING: Incomplete upgrade detected:")
		for _, node := range driftedNodes {
			fmt.Printf("  - %s is running kubelet %s and is not cordoned\n", node.Nodename, node.KubeletVersion.String())
		}
	}

	return nil
}
