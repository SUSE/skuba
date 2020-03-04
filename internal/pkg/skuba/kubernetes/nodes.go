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

package kubernetes

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubectldrain "k8s.io/kubernetes/pkg/kubectl/drain"
)

// GetAllNodes returns the list of nodes
func GetAllNodes(client clientset.Interface) (*corev1.NodeList, error) {
	return client.CoreV1().Nodes().List(metav1.ListOptions{})
}

// GetControlPlaneNodes returns the list of master nodes by matching
// "node-role.kubernetes.io/master" label.
func GetControlPlaneNodes(client clientset.Interface) (*corev1.NodeList, error) {
	return client.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=", kubeadmconstants.LabelNodeRoleMaster),
	})
}

// GetNodeWithMachineID returns the node matching machine ID.
func GetNodeWithMachineID(client clientset.Interface, machineID string) (*corev1.Node, error) {
	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range nodes.Items {
		if node.Status.NodeInfo.MachineID == machineID {
			return &node, nil
		}
	}
	return nil, errors.Errorf("node with machine-id %s not found", machineID)
}

// IsControlPlane check if given node is master node.
func IsControlPlane(node *corev1.Node) bool {
	_, isControlPlane := node.ObjectMeta.Labels[kubeadmconstants.LabelNodeRoleMaster]
	return isControlPlane
}

// DrainNode cordons, drains and evict given node.
func DrainNode(client clientset.Interface, node *corev1.Node, drainTimeout time.Duration) error {
	policyGroupVersion, err := kubectldrain.CheckEvictionSupport(client)
	if err != nil {
		return errors.Wrap(err, "could not get policy group version")
	}

	newCordon := kubectldrain.NewCordonHelper(node)
	newCordon.UpdateIfRequired(true)
	err, patchErr := newCordon.PatchOrReplace(client)
	if err != nil {
		return errors.Wrap(err, "failed to update node status")
	}
	if patchErr != nil {
		return errors.Wrap(patchErr, "failed to patch node status")
	}

	drainer := &kubectldrain.Helper{
		Client:              client,
		Force:               true,
		IgnoreAllDaemonSets: true,
		DeleteLocalData:     true,
		Timeout:             drainTimeout,
	}
	del, errs := drainer.GetPodsForDeletion(node.ObjectMeta.Name)
	if errs != nil {
		return fmt.Errorf("could not get pods for deletion %s", errs)
	}

	if len(policyGroupVersion) > 0 {
		for _, pod := range del.Pods() {
			if err = drainer.EvictPod(pod, policyGroupVersion); err != nil {
				return errors.Wrapf(err, "failed to evict pod: %v", pod.Name)
			}
		}
	} else {
		for _, pod := range del.Pods() {
			if err = drainer.DeletePod(pod); err != nil {
				return errors.Wrapf(err, "failed to delete pod: %v", pod.Name)
			}
		}
	}

	klog.V(1).Infof("node %s correctly drained", node.ObjectMeta.Name)

	return nil
}
