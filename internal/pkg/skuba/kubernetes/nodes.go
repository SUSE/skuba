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

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/pkg/errors"
)

type NodeVersionInfo struct {
	Nodename                 string
	ContainerRuntimeVersion  string
	KubeletVersion           *version.Version
	APIServerVersion         *version.Version
	ControllerManagerVersion *version.Version
	SchedulerVersion         *version.Version
	EtcdVersion              *version.Version
	Unschedulable            bool
}

func GetMasterNodes() (*v1.NodeList, error) {
	clientSet, err := GetAdminClientSet()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get admin clinet set")
	}
	return clientSet.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=", kubeadmconstants.LabelNodeRoleMaster),
	})
}

func IsMaster(node *v1.Node) bool {
	_, isMaster := node.ObjectMeta.Labels[kubeadmconstants.LabelNodeRoleMaster]
	return isMaster
}

func DrainNode(node *v1.Node) error {
	// Drain node (shelling out, FIXME after https://github.com/kubernetes/kubernetes/pull/72827 can be used [1.14])
	cmd := exec.Command("kubectl",
		fmt.Sprintf("--kubeconfig=%s", skuba.KubeConfigAdminFile()),
		"drain", "--delete-local-data=true", "--force=true", "--ignore-daemonsets=true", node.ObjectMeta.Name)

	if err := cmd.Run(); err != nil {
		klog.V(1).Infof("could not drain node %s, aborting (use --force if you want to ignore this error)", node.ObjectMeta.Name)
		return err
	} else {
		klog.V(1).Infof("node %s correctly drained", node.ObjectMeta.Name)
	}

	return nil
}

// AllNodesVersioningInfo returns the version info for all nodes in the cluster
func AllNodesVersioningInfo() ([]NodeVersionInfo, error) {
	client, err := GetAdminClientSet()
	if err != nil {
		return []NodeVersionInfo{}, errors.Wrap(err, "unable to get admin client set")
	}

	nodeList, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return []NodeVersionInfo{}, errors.Wrap(err, "could not retrieve node list")
	}

	result := []NodeVersionInfo{}

	for _, node := range nodeList.Items {
		nodeVersion, err := nodeVersioningInfoWithClientset(client, node.ObjectMeta.Name)
		if err != nil {
			return []NodeVersionInfo{}, err
		}
		result = append(result, nodeVersion)
	}

	return result, nil
}

// NodeVersioningInfo returns related versioning information about a node
func NodeVersioningInfo(nodeName string) (NodeVersionInfo, error) {
	client, err := GetAdminClientSet()
	if err != nil {
		return NodeVersionInfo{}, errors.Wrap(err, "unable to get admin client set")
	}

	nodeVersions, err := nodeVersioningInfoWithClientset(client, nodeName)
	if err != nil {
		return NodeVersionInfo{}, errors.Wrap(err, "unable to get node versioning info")
	}

	return nodeVersions, nil
}

func nodeVersioningInfoWithClientset(client kubernetes.Interface, nodeName string) (NodeVersionInfo, error) {
	nodeObject, err := client.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		return NodeVersionInfo{}, errors.Wrap(err, "could not retrieve node object")
	}

	kubeletVersion := version.MustParseSemantic(nodeObject.Status.NodeInfo.KubeletVersion)
	containerRuntimeVersion := nodeObject.Status.NodeInfo.ContainerRuntimeVersion
	unschedulable := nodeObject.Spec.Unschedulable

	nodeVersions := NodeVersionInfo{
		Nodename:                nodeName,
		ContainerRuntimeVersion: containerRuntimeVersion,
		KubeletVersion:          kubeletVersion,
		Unschedulable:           unschedulable,
	}

	// find out the container image tags, depending on the role of the node
	if IsMaster(nodeObject) {
		apiServerTag, _ := getPodContainerImageTagWithClientset(client, "kube-system", fmt.Sprintf("%s-%s", "kube-apiserver", nodeName))
		controllerManagerTag, _ := getPodContainerImageTagWithClientset(client, "kube-system", fmt.Sprintf("%s-%s", "kube-controller-manager", nodeName))
		schedulerTag, _ := getPodContainerImageTagWithClientset(client, "kube-system", fmt.Sprintf("%s-%s", "kube-scheduler", nodeName))
		etcdTag, _ := getPodContainerImageTagWithClientset(client, "kube-system", fmt.Sprintf("%s-%s", "etcd", nodeName))

		nodeVersions.APIServerVersion = version.MustParseSemantic(apiServerTag)
		nodeVersions.ControllerManagerVersion = version.MustParseSemantic(controllerManagerTag)
		nodeVersions.SchedulerVersion = version.MustParseSemantic(schedulerTag)
		nodeVersions.EtcdVersion = version.MustParseSemantic(etcdTag)
	}

	return nodeVersions, nil
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
