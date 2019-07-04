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

package remove

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/SUSE/skuba/internal/pkg/skuba/cni"
	"github.com/SUSE/skuba/internal/pkg/skuba/etcd"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/pkg/errors"
)

// Remove removes a node from the cluster
func Remove(target string, drainTimeout time.Duration) error {
	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "unable to get admin client set")
	}

	node, err := client.CoreV1().Nodes().Get(target, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "[remove-node] could not get node %s", target)
	}

	targetName := node.ObjectMeta.Name

	var isControlPlane bool
	if isControlPlane = kubernetes.IsControlPlane(node); isControlPlane {
		fmt.Printf("[remove-node] removing control plane node %s (drain timeout: %s)\n", targetName, drainTimeout.String())
	} else {
		fmt.Printf("[remove-node] removing worker node %s (drain timeout: %s)\n", targetName, drainTimeout.String())
	}

	kubernetes.DrainNode(node, drainTimeout)

	if isControlPlane {
		fmt.Printf("[remove-node] removing etcd from node %s\n", targetName)
		etcd.RemoveMember(node)
	}

	if err := kubernetes.DisarmKubelet(node); err != nil {
		fmt.Printf("[remove-node] failed disarming kubelet: %v; node could be down, continuing with node removal...\n", err)
	}

	if isControlPlane {
		if err := kubeadm.RemoveAPIEndpointFromConfigMap(node); err != nil {
			return errors.Wrapf(err, "[remove-node] could not remove the APIEndpoint for %s from the kubeadm-config configmap", targetName)
		}

		if err := cni.CreateOrUpdateCiliumConfigMap(); err != nil {
			return errors.Wrap(err, "[remove-node] could not update cilium-config configmap")
		}

		if err := cni.AnnotateCiliumDaemonsetWithCurrentTimestamp(); err != nil {
			fmt.Printf("[remove-node] could not annonate cilium daemonset: %v\n", err)
		}
	}

	if err := client.CoreV1().Nodes().Delete(targetName, &metav1.DeleteOptions{}); err != nil {
		errors.Wrapf(err, "[remove-node] could not remove node %s", targetName)
	}

	fmt.Printf("[remove-node] node %s successfully removed from the cluster\n", targetName)

	return nil
}
