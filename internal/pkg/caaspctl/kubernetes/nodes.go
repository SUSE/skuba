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

package kubernetes

import (
	"fmt"
	"os/exec"

	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/caaspctl/pkg/caaspctl"
	"github.com/pkg/errors"
)

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
		fmt.Sprintf("--kubeconfig=%s", caaspctl.KubeConfigAdminFile()),
		"drain", "--delete-local-data=true", "--force=true", "--ignore-daemonsets=true", node.ObjectMeta.Name)

	if err := cmd.Run(); err != nil {
		klog.V(1).Infof("could not drain node %s, aborting (use --force if you want to ignore this error)", node.ObjectMeta.Name)
		return err
	} else {
		klog.V(1).Infof("node %s correctly drained", node.ObjectMeta.Name)
	}

	return nil
}
