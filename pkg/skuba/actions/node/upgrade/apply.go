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

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/kured"
	"github.com/SUSE/skuba/internal/pkg/skuba/node"
	upgradenode "github.com/SUSE/skuba/internal/pkg/skuba/upgrade/node"
	"github.com/SUSE/skuba/pkg/skuba"
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

	// Check if a lock on kured already exists
	kuredWasLocked, err := kured.LockExists(client)
	if err != nil {
		return err
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
		if currentClusterVersion.Minor() == 17 {
			initCfg.UseHyperKubeImage = false
		}

		initCfgContents, err = kubeadm.UpdateClusterConfigurationWithClusterVersion(initCfg, nodeVersionInfoUpdate.Update.APIServerVersion)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Performing node %s (%s) upgrade, please wait...\n", target.Nodename, target.Target)

	if skubaUpdateWasEnabled {
		err = target.Apply(nil, "skuba-update-timer.disable")
		if err != nil {
			return err
		}
	}
	if !kuredWasLocked {
		if err := kured.Lock(client); err != nil {
			return err
		}
	}

	// Always upload crio files, regardless of the version (allows to
	// enforce user behavior during patch updates).
	// During the upgrade from 1.16 to 1.18 crio, the cri-o package will
	// handle overriding the old crio sysconfig to an "empty" sysconfig,
	// and the cri.configure action will be enough.
	// We can remove the conditionals and only
	// keep the cri.configure action when caasp 4.2.0 is not supported
	// anymore (everyone has updated crio to 1.18)
	if _, err := os.Stat(skuba.CriDefaultsConfFile()); err == nil {
		err = target.Apply(nil, "cri.configure")
		if err != nil {
			return err
		}
	} else if _, err := os.Stat(skuba.CriDockerDefaultsConfFile()); err == nil {
		err = target.Apply(nil, "cri.sysconfig")
		if err != nil {
			return err
		}
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
		"kubernetes.restart-services",
	)
	if err != nil {
		return err
	}
	// After the node upgrade to Kubernetes 1.18, Cilium pods are stuck in
	// the Init:Error state due to cri-o issue;
	// "Error reserving pod name k8s_cilium-<endofpodname>kube-system<number> for id <id>: name is reserved"
	// Deleting the pod will trigger its creation by k8s, without problem
	// of name reservation. To avoid a massive flood of deletion, we only
	// select the Cilium pod running on the node on which we have done the
	// upgrade.
	if currentVersion == "1.17" {
		ciliumPod, err := kubernetes.GetFirstPodMatchingSelectors(
			client,
			metav1.NamespaceSystem,
			"k8s-app=cilium",
			fmt.Sprintf("spec.nodeName=%s", target.Nodename),
		)
		if err != nil {
			return err
		}
		switch ciliumPod.Status.Phase {
		case v1.PodFailed:
			klog.V(4).Info("cilium pod in failed state")
			if err := kubernetes.DeletePod(client, metav1.NamespaceSystem, ciliumPod.GetName()); err != nil {
				return errors.Wrap(err, "could not delete failed Cilium pod")
			}
		case v1.PodUnknown:
			// TODO: rmeove this case if not needed
			klog.V(4).Info("cilium pod in unknown state")
			if err := kubernetes.DeletePod(client, metav1.NamespaceSystem, ciliumPod.GetName()); err != nil {
				return errors.Wrap(err, "could not delete failed Cilium pod")
			}
		default:
			klog.V(4).Infof("cilium pod state: %s", ciliumPod.Status.Phase)
		}
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
