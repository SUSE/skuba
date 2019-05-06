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
	"fmt"
	"io/ioutil"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	kubeadmtokenphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/bootstraptoken/node"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/kubernetes"
	"github.com/SUSE/caaspctl/pkg/caaspctl"
	"github.com/pkg/errors"
)

// Join joins a new machine to the cluster. The role of the machine will be
// provided by the JoinConfiguration, and will target Target node
//
// FIXME: being this a part of the go API accept the toplevel directory instead of
//        using the PWD
func Join(joinConfiguration deployments.JoinConfiguration, target *deployments.Target) error {
	statesToApply := []string{"kernel.load-modules", "kernel.configure-parameters",
		"cri.start", "kubelet.configure", "kubelet.enable", "kubeadm.join", "cni.cilium-update-configmap"}

	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		fmt.Print("[join] failed to get admin client set\n")
		return err
	}

	_, err = client.CoreV1().Nodes().Get(target.Nodename, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("[join] failed to join the node with name %q since a node with the same name already exists in the cluster\n", target.Nodename)
		return err
	}

	if joinConfiguration.Role == deployments.MasterRole {
		statesToApply = append([]string{"kubernetes.join.upload-secrets"}, statesToApply...)
	}

	fmt.Println("[join] applying states to new node")

	if err := target.Apply(joinConfiguration, statesToApply...); err != nil {
		fmt.Printf("[join] failed to apply join to node %s\n", err)
		return err
	}

	fmt.Println("[join] node successfully joined the cluster")
	return nil
}

// ConfigPath returns the configuration path for a specific Target; if this file does
// not exist, it will be created out of the template file
//
// FIXME: being this a part of the go API accept the toplevel directory instead of
//        using the PWD
func ConfigPath(role deployments.Role, target *deployments.Target) (string, error) {
	configPath := caaspctl.MachineConfFile(target.Target)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = caaspctl.TemplatePathForRole(role)
	}

	joinConfiguration, err := LoadJoinConfigurationFromFile(configPath)
	if err != nil {
		return "", errors.Wrap(err, "error parsing configuration")
	}
	addFreshTokenToJoinConfiguration(target.Target, joinConfiguration)
	addTargetInformationToJoinConfiguration(target, role, joinConfiguration)
	finalJoinConfigurationContents, err := kubeadmconfigutil.MarshalKubeadmConfigObject(joinConfiguration)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal configuration")
	}

	if err := ioutil.WriteFile(caaspctl.MachineConfFile(target.Target), finalJoinConfigurationContents, 0600); err != nil {
		return "", errors.Wrap(err, "error writing specific machine configuration")
	}

	return caaspctl.MachineConfFile(target.Target), nil
}

func addFreshTokenToJoinConfiguration(target string, joinConfiguration *kubeadmapi.JoinConfiguration) error {
	if joinConfiguration.Discovery.BootstrapToken == nil {
		joinConfiguration.Discovery.BootstrapToken = &kubeadmapi.BootstrapTokenDiscovery{}
	}
	var err error
	joinConfiguration.Discovery.BootstrapToken.Token, err = createBootstrapToken(target)
	joinConfiguration.Discovery.TLSBootstrapToken = ""
	return err
}

func addTargetInformationToJoinConfiguration(target *deployments.Target, role deployments.Role, joinConfiguration *kubeadmapi.JoinConfiguration) error {
	if joinConfiguration.NodeRegistration.KubeletExtraArgs == nil {
		joinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
	}
	joinConfiguration.NodeRegistration.Name = target.Nodename
	joinConfiguration.NodeRegistration.CRISocket = caaspctl.CRISocket
	joinConfiguration.NodeRegistration.KubeletExtraArgs["hostname-override"] = target.Nodename
	joinConfiguration.NodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = images.GetGenericImage(caaspctl.ImageRepository, "pause", kubernetes.CurrentComponentVersion(kubernetes.Pause))
	isSUSE, err := target.IsSUSEOS()
	if err != nil {
		return errors.Wrap(err, "unable to get os info")
	}
	if isSUSE {
		joinConfiguration.NodeRegistration.KubeletExtraArgs["cni-bin-dir"] = caaspctl.SUSECNIDir
	}
	return nil
}

func createBootstrapToken(target string) (string, error) {
	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return "", errors.Wrap(err, "unable to get admin client set")
	}

	bootstrapTokenRaw, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return "", errors.Wrap(err, "could not generate a new bootstrap token")
	}

	bootstrapToken, err := kubeadmapi.NewBootstrapTokenString(bootstrapTokenRaw)
	if err != nil {
		return "", errors.Wrap(err, "could not generate a new boostrap token")
	}

	bootstrapTokens := []kubeadmapi.BootstrapToken{
		{
			Token:       bootstrapToken,
			Description: fmt.Sprintf("Bootstrap token for %s machine join", target),
			TTL: &metav1.Duration{
				Duration: 15 * time.Minute,
			},
			Usages: []string{"signing", "authentication"},
			Groups: []string{"system:bootstrappers:kubeadm:default-node-token"},
		},
	}

	if err := kubeadmtokenphase.CreateNewTokens(client, bootstrapTokens); err != nil {
		return "", errors.Wrap(err, "could not create new bootstrap token")
	}

	return bootstrapTokenRaw, nil
}
