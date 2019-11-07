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

package node

import (
	"k8s.io/apimachinery/pkg/util/version"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

func AddTargetInformationToInitConfigurationWithClusterVersion(target *deployments.Target, initConfiguration *kubeadmapi.InitConfiguration, clusterVersion *version.Version) error {
	if initConfiguration.NodeRegistration.KubeletExtraArgs == nil {
		initConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
	}
	initConfiguration.NodeRegistration.Name = target.Nodename
	initConfiguration.NodeRegistration.CRISocket = skuba.CRISocket
	initConfiguration.NodeRegistration.KubeletExtraArgs["hostname-override"] = target.Nodename
	initConfiguration.NodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = kubernetes.ComponentContainerImageForClusterVersion(kubernetes.Pause, clusterVersion)
	isSUSE, err := target.IsSUSEOS()
	if err != nil {
		return err
	}
	if isSUSE {
		initConfiguration.NodeRegistration.KubeletExtraArgs["cni-bin-dir"] = skuba.SUSECNIDir
	}
	return nil
}
