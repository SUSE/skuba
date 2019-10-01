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
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/kured"
	"github.com/SUSE/skuba/internal/pkg/skuba/node"
	upgradenode "github.com/SUSE/skuba/internal/pkg/skuba/upgrade/node"
	"github.com/pkg/errors"
)

func Apply(target *deployments.Target) error {
	if err := fillTargetWithNodeName(target); err != nil {
		return err
	}

	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return err
	}

	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
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

	// Check if skuba-update.timer is already disabled
	skubaUpdateWasEnabled, err := target.IsServiceEnabled("skuba-update.timer")
	if err != nil {
		return err
	}

	// Check if a lock on kured already exists
	kuredWasLocked, err := kured.LockExists(client)
	if err != nil {
		return err
	}

	var upgradeable bool
	var initCfgContents []byte

	// Check if the target node is the first control plane to be updated
	isFirstControlPlaneUpgrade, err := nodeVersionInfoUpdate.IsFirstControlPlaneNodeToBeUpgraded()
	if err != nil {
		return err
	}
	if isFirstControlPlaneUpgrade {
		var err error
		upgradeable, err = kubernetes.AllWorkerNodesTolerateVersion(nodeVersionInfoUpdate.Update.APIServerVersion)
		if err != nil {
			return err
		}
		if upgradeable {
			fmt.Println("Fetching the cluster configuration...")

			initCfg, err := kubeadm.GetClusterConfiguration(client)
			if err != nil {
				return err
			}
			node.AddTargetInformationToInitConfigurationWithClusterVersion(target, initCfg, nodeVersionInfoUpdate.Update.APIServerVersion)
			kubeadm.SetContainerImagesWithClusterVersion(initCfg, nodeVersionInfoUpdate.Update.APIServerVersion)
			initCfgContents, err = kubeadmconfigutil.MarshalInitConfigurationToBytes(initCfg, schema.GroupVersion{
				Group:   "kubeadm.k8s.io",
				Version: "v1beta2",
			})
			if err != nil {
				return err
			}
		}
	} else {
		// there is already at least one updated control plane node
		if nodeVersionInfoUpdate.Current.IsControlPlane() {
			upgradeable, err = kubernetes.AllWorkerNodesTolerateVersion(currentClusterVersion)
			if err != nil {
				return err
			}
		} else {
			// worker nodes have no preconditions, are always upgradeable
			upgradeable = true
		}
	}

	if !upgradeable {
		return errors.Errorf("node %s cannot be upgraded yet", target.Nodename)
	}

	fmt.Printf("Performing node %s (%s) upgrade, please wait...\n", target.Nodename, target.Target)

	if skubaUpdateWasEnabled {
		err = target.Apply(nil, "skuba-update.stop")
		if err != nil {
			return err
		}
	}
	if !kuredWasLocked {
		if err := kured.Lock(client); err != nil {
			return err
		}
	}
	if nodeVersionInfoUpdate.HasMajorOrMinorUpdate() {
		err = target.Apply(deployments.KubernetesBaseOSConfiguration{
			UpdatedVersion: nodeVersionInfoUpdate.Update.KubeletVersion.String(),
			CurrentVersion: nodeVersionInfoUpdate.Current.KubeletVersion.String(),
		}, "kubernetes.install-node-pattern")
		if err != nil {
			return err
		}
	}
	if isFirstControlPlaneUpgrade {
		err = target.Apply(deployments.UpgradeConfiguration{
			KubeadmConfigContents: string(initCfgContents),
		}, "kubeadm.upgrade.apply")
		if err != nil {
			return err
		}
	} else if err := target.Apply(nil, "kubeadm.upgrade.node"); err != nil {
		return err
	}
	err = target.Apply(deployments.KubernetesBaseOSConfiguration{
		CurrentVersion: nodeVersionInfoUpdate.Update.KubeletVersion.String(),
	}, "kubernetes.install-node-pattern")
	if err != nil {
		return err
	}
	if err := target.Apply(nil, "kubernetes.restart-services"); err != nil {
		return err
	}
	if err := target.Apply(nil, "kubernetes.wait-for-kubelet"); err != nil {
		klog.Errorf("Kubelet could not register node %s. Please check the kubelet system logs and be aware that services kured or skuba-update will stay disabled", target.Nodename)
		return err
	}
	if skubaUpdateWasEnabled {
		err = target.Apply(nil, "skuba-update.start")
		if err != nil {
			return err
		}
	}
	if !kuredWasLocked {
		if err := kured.Unlock(client); err != nil {
			return err
		}
	}

	fmt.Printf("Node %s (%s) successfully upgraded\n", target.Nodename, target.Target)

	return nil
}

func fillTargetWithNodeName(target *deployments.Target) error {
	machineId, err := target.DownloadFileContents("/etc/machine-id")
	if err != nil {
		return err
	}
	node, err := kubernetes.GetNodeWithMachineId(strings.TrimSuffix(machineId, "\n"))
	if err != nil {
		return err
	}
	target.Nodename = node.ObjectMeta.Name
	return nil
}
