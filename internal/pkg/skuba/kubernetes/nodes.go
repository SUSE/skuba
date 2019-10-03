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
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/pkg/skuba"
)

func GetControlPlaneNodes(client kubernetes.Interface) (*v1.NodeList, error) {
	return client.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=", kubeadmconstants.LabelNodeRoleMaster),
	})
}

func GetNodeWithMachineId(machineId string) (*v1.Node, error) {
	clientSet, err := GetAdminClientSet()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get admin client set")
	}
	nodes, err := clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, node := range nodes.Items {
		if node.Status.NodeInfo.MachineID == machineId {
			return &node, nil
		}
	}
	return nil, errors.Errorf("node with machine-id %s not found", machineId)
}

func IsControlPlane(node *v1.Node) bool {
	_, isControlPlane := node.ObjectMeta.Labels[kubeadmconstants.LabelNodeRoleMaster]
	return isControlPlane
}

func DrainNode(node *v1.Node, drainTimeout time.Duration) error {
	// Drain node (shelling out, FIXME after https://github.com/kubernetes/kubernetes/pull/72827 can be used [1.14])
	cmd := exec.Command("kubectl",
		fmt.Sprintf("--kubeconfig=%s", skuba.KubeConfigAdminFile()),
		"drain", "--delete-local-data=true", "--force=true", "--ignore-daemonsets=true", fmt.Sprintf("--timeout=%s", drainTimeout.String()), node.ObjectMeta.Name)

	if err := cmd.Run(); err != nil {
		klog.V(1).Infof("could not drain node %s, aborting (use --force if you want to ignore this error)", node.ObjectMeta.Name)
		return err
	} else {
		klog.V(1).Infof("node %s correctly drained", node.ObjectMeta.Name)
	}

	return nil
}

func getPodContainerImageTagWithClientset(client kubernetes.Interface, namespace string, podName string) (string, error) {
	podObject, err := client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "could not retrieve pod object")
	}
	containerImageWithName := podObject.Spec.Containers[0].Image
	containerImageTag := strings.Split(containerImageWithName, ":")

	return containerImageTag[len(containerImageTag)-1], nil
}
