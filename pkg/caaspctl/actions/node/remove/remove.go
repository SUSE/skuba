/*
 * Copyright (c) 2019 SUSE LLC. All rights reserved.
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

package node

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/cni"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/etcd"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/kubeadm"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/kubernetes"
)

// Remove removes a node from the cluster
//
// FIXME: error handling with `github.com/pkg/errors`; return errors
func Remove(target string, isForced bool) {
	client := kubernetes.GetAdminClientSet()

	node, err := client.CoreV1().Nodes().Get(target, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("[remove-node] could not get node %s: %v\n", target, err)
		os.Exit(1)
	}

	targetName := node.ObjectMeta.Name
	isMaster := kubernetes.IsMaster(node)

	masterNodes, err := kubernetes.GetMasterNodes()
	if err != nil {
		fmt.Printf("[remove-node] could not get list of master nodes %s\n", err)
		return
	}

	if isMaster && !isForced {
		if len(masterNodes.Items) == 1 {
			fmt.Println("[remove-node] WARNING: trying to remove last master node from the cluster")
			fmt.Print("[remove-node] Are you sure you want to proceed? [y/N]: ")
			s := bufio.NewScanner(os.Stdin)
			s.Scan()
			if err := s.Err(); err != nil {
				fmt.Println(err)
				return
			}
			if strings.ToLower(s.Text()) != "y" {
				fmt.Println("Aborted remove-node operation")
				return
			}
		}
	}

	if isMaster {
		fmt.Printf("[remove-node] removing master node %s\n", targetName)
	} else {
		fmt.Printf("[remove-node] removing worker node %s\n", targetName)
	}

	kubernetes.DrainNode(node)

	if isMaster {
		fmt.Printf("[remove-node] removing etcd from node %s\n", targetName)
		etcd.RemoveMember(node)
	}

	if err := kubernetes.DisarmKubelet(node); err != nil {
		fmt.Printf("[remove-node] failed disarming kubelet: %v; node could be down, continuing with node removal...", err)
	}

	if isMaster {
		if err := kubeadm.RemoveAPIEndpointFromConfigMap(node); err != nil {
			fmt.Printf("[remove-node] could not remove the APIEndpoint for %s from the kubeadm-config configmap", targetName)
		} else {
			if err := cni.CreateOrUpdateCiliumConfigMap(); err != nil {
				fmt.Printf("[remove-node] could not update cilium-config configmap: %v", err)
			} else {
				if err := cni.AnnotateCiliumDaemonsetWithCurrentTimestamp(); err != nil {
					fmt.Printf("[remove-node] could not annonate cilium daemonset: %v", err)
				}
			}
		}
	}

	if err := client.CoreV1().Nodes().Delete(targetName, &metav1.DeleteOptions{}); err == nil {
		fmt.Printf("[remove-node] node %s successfully removed from the cluster\n", targetName)
	} else {
		fmt.Printf("[remove-node] could not remove node %s\n", targetName)
	}
}
