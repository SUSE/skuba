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

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	upgradenode "github.com/SUSE/skuba/internal/pkg/skuba/upgrade/node"
	"github.com/SUSE/skuba/pkg/skuba"
)

func Apply(target *deployments.Target) error {
	fmt.Printf("%s\n", skuba.CurrentVersion().String())

	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion()
	if err != nil {
		return err
	}
	currentVersion := currentClusterVersion.String()
	latestVersion := kubernetes.LatestVersion().String()
	fmt.Printf("Current Kubernetes cluster version: %s\n", currentVersion)
	fmt.Printf("Latest Kubernetes version: %s\n", latestVersion)
	fmt.Println()

	nodeVersionInfoUpdate, err := upgradenode.UpdateStatus(target.Nodename)
	if err != nil {
		return err
	}

	if nodeVersionInfoUpdate.IsUpdated() {
		fmt.Printf("Node %s is up to date\n", target.Nodename)
		return nil
	}

	// Check If the target node is the first control plane to be updated
	if nodeVersionInfoUpdate.IsFirstControlPlaneNodeToBeUpgraded() {
		upgradeable, err := kubernetes.AllWorkerNodesTolerateUpdate()
		if err != nil {
			return err
		}
		if upgradeable {
			fmt.Println("TODO: trigger update installation")
		}
	} else {
		// there is already at least one updated control plane node
		upgradeable := true
		if nodeVersionInfoUpdate.Current.IsControlPlane() {
			upgradeable, err = kubernetes.AllWorkerNodesTolerateUpdate()
			if err != nil {
				return err
			}
		}
		if upgradeable {
			fmt.Println("TODO: trigger update installation")
		}
	}

	return nil
}
