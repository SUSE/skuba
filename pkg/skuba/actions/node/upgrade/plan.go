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

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	upgradenode "github.com/SUSE/skuba/internal/pkg/skuba/upgrade/node"
	clientset "k8s.io/client-go/kubernetes"
)

func Plan(client clientset.Interface, nodeName string) error {
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return err
	}
	nodeVersionInfoUpdate, err := upgradenode.UpdateStatus(client, nodeName)
	if err != nil {
		return err
	}

	fmt.Printf("Current Kubernetes cluster version: %s\n", currentClusterVersion.String())
	fmt.Printf("Latest Kubernetes version: %s\n", kubernetes.LatestVersion().String())
	fmt.Printf("Current Node version: %s\n", nodeVersionInfoUpdate.Current.KubeletVersion.String())
	fmt.Println()

	if nodeVersionInfoUpdate.IsUpdated() {
		fmt.Printf("Node %s is up to date\n", nodeName)
	} else {
		fmt.Printf("Component versions in %s\n", nodeName)
		if nodeVersionInfoUpdate.Current.IsControlPlane() {
			fmt.Printf("  - apiserver: %s -> %s\n", nodeVersionInfoUpdate.Current.APIServerVersion.String(), nodeVersionInfoUpdate.Update.APIServerVersion.String())
			fmt.Printf("  - controller-manager: %s -> %s\n", nodeVersionInfoUpdate.Current.ControllerManagerVersion.String(), nodeVersionInfoUpdate.Update.ControllerManagerVersion.String())
			fmt.Printf("  - scheduler: %s -> %s\n", nodeVersionInfoUpdate.Current.SchedulerVersion.String(), nodeVersionInfoUpdate.Update.SchedulerVersion.String())
			fmt.Printf("  - etcd: %s -> %s\n", nodeVersionInfoUpdate.Current.EtcdVersion.String(), nodeVersionInfoUpdate.Update.EtcdVersion.String())
		}
		fmt.Printf("  - kubelet: %s -> %s\n", nodeVersionInfoUpdate.Current.KubeletVersion.String(), nodeVersionInfoUpdate.Update.KubeletVersion.String())
		fmt.Printf("  - cri-o: %s -> %s\n", nodeVersionInfoUpdate.Current.ContainerRuntimeVersion.String(), nodeVersionInfoUpdate.Update.ContainerRuntimeVersion.String())

		// Check if the node is upgradeable (matches preconditions)
		if err := nodeVersionInfoUpdate.NodeUpgradeableCheck(client, currentClusterVersion); err != nil {
			fmt.Println()
			return err
		}
	}

	return nil
}
