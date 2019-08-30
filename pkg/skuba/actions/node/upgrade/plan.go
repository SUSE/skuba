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
	"github.com/SUSE/skuba/pkg/skuba"
)

func Plan(nodeName string) error {
	fmt.Printf("%s\n", skuba.CurrentVersion().String())

	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return err
	}
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return err
	}
	nodeVersionInfoUpdate, err := upgradenode.UpdateStatus(nodeName)
	if err != nil {
		return err
	}

	fmt.Printf("Current Kubernetes cluster version: %s\n", currentClusterVersion.String())
	fmt.Printf("Latest Kubernetes version: %s\n", kubernetes.LatestVersion().String())
	fmt.Println()
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
	}

	return nil
}
