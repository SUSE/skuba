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

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/internal/pkg/skuba/cni"
	"github.com/SUSE/skuba/internal/pkg/skuba/etcd"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

// Remove removes a node from the cluster
func Remove(client clientset.Interface, target string, drainTimeout time.Duration) error {
	node, err := client.CoreV1().Nodes().Get(target, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "[remove-node] could not get node %s", target)
	}

	if kubernetes.IsControlPlane(node) {
		nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=", kubeadmconstants.LabelNodeRoleMaster),
		})

		if err != nil {
			return errors.Wrapf(err, "could not retrieve master node list")
		}

		if len(nodes.Items) == 1 {
			return errors.New("could not remove last master of the cluster")
		}
	}

	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return errors.Wrap(err, "could not retrieve the current cluster version")
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
		etcd.RemoveMember(client, node)
	}

	if err := kubernetes.DisarmKubelet(client, node, currentClusterVersion); err != nil {
		fmt.Printf("[remove-node] failed disarming kubelet: %v; node could be down, continuing with node removal...\n", err)
	}

	if isControlPlane {
		if err := kubeadm.RemoveAPIEndpointFromConfigMap(client, node); err != nil {
			return errors.Wrapf(err, "[remove-node] could not remove the APIEndpoint for %s from the kubeadm-config configmap", targetName)
		}

		if err := cni.CreateOrUpdateCiliumConfigMap(client); err != nil {
			return errors.Wrap(err, "[remove-node] could not update cilium-config configmap")
		}
	}

	if err := client.CoreV1().Nodes().Delete(targetName, &metav1.DeleteOptions{}); err != nil {
		errors.Wrapf(err, "[remove-node] could not remove node %s", targetName)
	}

	fmt.Printf("[remove-node] node %s successfully removed from the cluster\n", targetName)

	return nil
}
