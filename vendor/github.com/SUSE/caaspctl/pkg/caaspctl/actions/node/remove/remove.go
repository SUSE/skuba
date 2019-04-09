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
	"k8s.io/klog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/etcd"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/kubeadm"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/kubernetes"
)

// Remove removes a node from the cluster
//
// FIXME: error handling with `github.com/pkg/errors`; return errors
func Remove(target string) {
	client := kubernetes.GetAdminClientSet()

	node, err := client.CoreV1().Nodes().Get(target, metav1.GetOptions{})
	if err != nil {
		klog.Fatalf("could not get node %s: %v\n", target, err)
	}

	targetName := node.ObjectMeta.Name

	var isMaster bool
	if isMaster = kubernetes.IsMaster(node); isMaster {
		klog.Infof("removing master node %s\n", targetName)
	} else {
		klog.Infof("removing worker node %s\n", targetName)
	}

	kubernetes.DrainNode(node)

	if isMaster {
		etcd.RemoveMember(node)
	}

	if err := kubernetes.DisarmKubelet(node); err != nil {
		klog.Infof("error disarming kubelet: %v; node could be down, continuing with node removal...", err)
	}

	if isMaster {
		if err := kubeadm.RemoveAPIEndpointFromConfigMap(node); err != nil {
			klog.Infof("could not remove the APIEndpoint for %s from the kubeadm-config configmap", targetName)
		}
	}

	if err := client.CoreV1().Nodes().Delete(targetName, &metav1.DeleteOptions{}); err == nil {
		klog.Infof("node %s successfully removed from the cluster\n", targetName)
	} else {
		klog.Infof("could not remove node %s\n", targetName)
	}
}
