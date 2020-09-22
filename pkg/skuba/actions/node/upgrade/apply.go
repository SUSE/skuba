/*
 * Copyright (c) 2019,2020 SUSE LLC.
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
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	clientset "k8s.io/client-go/kubernetes"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/kured"
	"github.com/SUSE/skuba/internal/pkg/skuba/node"
	upgradenode "github.com/SUSE/skuba/internal/pkg/skuba/upgrade/node"
	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/pkg/errors"
)

func Apply(client clientset.Interface, target *deployments.Target) error {
	if err := fillTargetWithNodeNameAndRole(client, target); err != nil {
		return err
	}

	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return err
	}
	currentVersion := currentClusterVersion.String()
	latestVersion := kubernetes.LatestVersion().String()
	nodeVersionInfoUpdate, err := upgradenode.UpdateStatus(client, target.Nodename)
	if err != nil {
		return err
	}

	fmt.Printf("Current Kubernetes cluster version: %s\n", currentVersion)
	fmt.Printf("Latest Kubernetes version: %s\n", latestVersion)
	fmt.Printf("Current Node version: %s\n", nodeVersionInfoUpdate.Current.KubeletVersion.String())
	fmt.Println()

	if nodeVersionInfoUpdate.IsUpdated() {
		fmt.Printf("Node %s is up to date\n", target.Nodename)
		return nil
	}

	// Refreshing cache info about target OS
	target.IsSUSEOS()
	// Check if the node is upgradeable (matches preconditions)
	if err := nodeVersionInfoUpdate.NodeUpgradeableCheck(client, currentClusterVersion); err != nil {
		fmt.Println()
		return err
	}

	// Check if skuba-update.timer is already disabled
	skubaUpdateWasEnabled, err := target.IsServiceEnabled("skuba-update.timer")
	if err != nil {
		return err
	}

	// Disable skuba-update.timer before upgrade
	if skubaUpdateWasEnabled {
		err = target.Apply(nil, "skuba-update-timer.disable")
		if err != nil {
			return err
		}
	}

	// Check if a kured reboot file already exists
	kuredRebootFilePresent := kured.RebootFileExists()
	if kuredRebootFilePresent {
		err := kured.RebootFileRemove()
		if err != nil {
			return err
		}
	}

	// Check if a lock on kured already exists
	kuredWasLocked, err := kured.LockExists(client)
	if err != nil {
		return err
	}

	// Lock kured before upgrade
	if !kuredWasLocked {
		if err := kured.Lock(client); err != nil {
			return err
		}
	}

	var initCfgContents []byte

	// Check if it's the first control plane node to be upgraded
	isFirstControlPlaneNodeToBeUpgraded, err := nodeVersionInfoUpdate.IsFirstControlPlaneNodeToBeUpgraded(client)
	if err != nil {
		return err
	}
	if isFirstControlPlaneNodeToBeUpgraded {
		fmt.Println("Fetching the cluster configuration...")

		initCfg, err := kubeadm.GetClusterConfiguration(client)
		if err != nil {
			return err
		}
		if err := node.AddTargetInformationToInitConfigurationWithClusterVersion(target, initCfg, nodeVersionInfoUpdate.Update.APIServerVersion); err != nil {
			return errors.Wrap(err, "error adding target information to init configuration")
		}

		// Upgrade 1.17 to 1.18.
		// This updated UseHyperKube field in-memory (unsets it).
		// Note: The cluster cm is uploaded at the end of the kubeadm process, as usual.
		// The whole paragraph can be removed when upgrading from 1.17 is removed.
		if currentClusterVersion.Minor() == 17 && kubernetes.LatestVersion().Minor() > 17 {
			initCfg.UseHyperKubeImage = false
		}

		kubeadm.UpdateClusterConfigurationWithClusterVersion(initCfg, nodeVersionInfoUpdate.Update.APIServerVersion)
		initCfgContents, err = kubeadmconfigutil.MarshalInitConfigurationToBytes(initCfg, schema.GroupVersion{
			Group:   "kubeadm.k8s.io",
			Version: kubeadm.GetKubeadmApisVersion(nodeVersionInfoUpdate.Update.APIServerVersion),
		})
		if err != nil {
			return err
		}
	}

	const drainTimeout = 15 * time.Minute
	node := nodeVersionInfoUpdate.Current.Node
	fmt.Printf("Draining node %s (timeout %dmin)\n", target.Nodename, drainTimeout)
	if err := kubernetes.DrainNode(client, node, drainTimeout); err != nil {
		return errors.Wrapf(err, "draining node %s", target.Nodename)
	}

	fmt.Printf("Performing node %s (%s) upgrade, please wait...\n", target.Nodename, target.Target)

	// Always upload crio files, regardless of the version (allows to enforce
	// user behavior during patch updates).
	if _, err := os.Stat(skuba.CriDefaultsConfFile()); err != nil {
		return errors.Wrap(err, "you need to migrate the local configuration of the cluster. Run: `skuba cluster upgrade localconfig`")
	}
	err = target.Apply(nil, "cri.configure", "cri.sysconfig")
	if err != nil {
		return err
	}

	if nodeVersionInfoUpdate.HasMajorOrMinorUpdate() {
		err = target.Apply(deployments.KubernetesBaseOSConfiguration{
			UpdatedVersion: nodeVersionInfoUpdate.Update.KubeletVersion.String(),
			CurrentVersion: nodeVersionInfoUpdate.Current.KubeletVersion.String(),
		}, "kubernetes.upgrade-stage-one")
		if err != nil {
			return err
		}
	}
	if isFirstControlPlaneNodeToBeUpgraded {
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
		UpdatedVersion: nodeVersionInfoUpdate.Update.KubeletVersion.String(),
		CurrentVersion: nodeVersionInfoUpdate.Current.KubeletVersion.String(),
	}, "kubernetes.upgrade-stage-two")
	if err != nil {
		return err
	}

	// bsc#1155810: generate cluster-wide kubelet root certificate, and generate/rotate kuberlet server certificate
	if err := kubernetes.GenerateKubeletRootCert(); err != nil {
		return err
	}
	err = target.Apply(nil,
		"kubelet.rootcert.upload",
		"kubelet.servercert.create-and-upload",
		"kubernetes.restart-services",
		"kubernetes.enable-services",
	)
	if err != nil {
		return err
	}

	if skubaUpdateWasEnabled {
		err = target.Apply(nil,
			"skuba-update.start.no-block",
			"skuba-update-timer.enable",
		)
		if err != nil {
			return err
		}
	}
	if !kuredWasLocked {
		if err := kured.Unlock(client); err != nil {
			return err
		}
	}
	if kuredRebootFilePresent {
		err := kured.RebootFileCreate()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Uncordon node %s\n", target.Nodename)
	if err := kubernetes.UncordonNode(client, node); err != nil {
		return errors.Wrapf(err, "uncordon node %s", target.Nodename)
	}

	fmt.Printf("Node %s (%s) successfully upgraded\n", target.Nodename, target.Target)

	return nil
}

func fillTargetWithNodeNameAndRole(client clientset.Interface, target *deployments.Target) error {
	machineID, err := target.DownloadFileContents("/etc/machine-id")
	if err != nil {
		return err
	}
	node, err := kubernetes.GetNodeWithMachineID(client, strings.TrimSuffix(machineID, "\n"))
	if err != nil {
		return err
	}
	target.Nodename = node.ObjectMeta.Name

	var role deployments.Role
	if kubernetes.IsControlPlane(node) {
		role = deployments.MasterRole
	} else {
		role = deployments.WorkerRole
	}
	target.Role = &role

	return nil
}
