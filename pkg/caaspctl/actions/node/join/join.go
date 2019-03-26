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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeadmtokenphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/bootstraptoken/node"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/config/strict"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
	"suse.com/caaspctl/internal/pkg/caaspctl/kubernetes"
	"suse.com/caaspctl/pkg/caaspctl"
)

// Join joins a new machine to the cluster. The role of the machine will be
// provided by the JoinConfiguration, and will target Target node
//
// FIXME: being this a part of the go API accept the toplevel directory instead of
//        using the PWD
// FIXME: error handling with `github.com/pkg/errors`; return errors
func Join(joinConfiguration deployments.JoinConfiguration, target *deployments.Target) {
	statesToApply := []string{"kubelet.configure", "kubelet.enable", "kubeadm.join"}

	if joinConfiguration.Role == deployments.MasterRole {
		statesToApply = append([]string{"kubernetes.join.upload-secrets"}, statesToApply...)
	}

	target.Apply(joinConfiguration, statesToApply...)
}

// ConfigPath returns the configuration path for a specific Target; if this file does
// not exist, it will be created out of the template file
//
// FIXME: being this a part of the go API accept the toplevel directory instead of
//        using the PWD
// FIXME: error handling with `github.com/pkg/errors`; return errors
func ConfigPath(role deployments.Role, target *deployments.Target) string {
	configPath := caaspctl.MachineConfFile(target.Target)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = caaspctl.TemplatePathForRole(role)
	}

	joinConfiguration, err := joinConfigFileAndDefaultsToInternalConfig(configPath)
	if err != nil {
		log.Fatalf("error parsing configuration: %v", err)
	}
	addFreshTokenToJoinConfiguration(target.Target, joinConfiguration)
	addTargetInformationToJoinConfiguration(target, role, joinConfiguration)
	finalJoinConfigurationContents, err := kubeadmconfigutil.MarshalKubeadmConfigObject(joinConfiguration)
	if err != nil {
		log.Fatal("could not marshal configuration")
	}

	if err := ioutil.WriteFile(caaspctl.MachineConfFile(target.Target), finalJoinConfigurationContents, 0600); err != nil {
		log.Fatal("error writing specific machine configuration")
	}

	return caaspctl.MachineConfFile(target.Target)
}

func addFreshTokenToJoinConfiguration(target string, joinConfiguration *kubeadmapi.JoinConfiguration) {
	if joinConfiguration.Discovery.BootstrapToken == nil {
		joinConfiguration.Discovery.BootstrapToken = &kubeadmapi.BootstrapTokenDiscovery{}
	}
	joinConfiguration.Discovery.BootstrapToken.Token = createBootstrapToken(target)
	joinConfiguration.Discovery.TLSBootstrapToken = ""
}

func addTargetInformationToJoinConfiguration(target *deployments.Target, role deployments.Role, joinConfiguration *kubeadmapi.JoinConfiguration) {
	if joinConfiguration.NodeRegistration.KubeletExtraArgs == nil {
		joinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
	}
	joinConfiguration.NodeRegistration.Name = target.Nodename
	joinConfiguration.NodeRegistration.KubeletExtraArgs["hostname-override"] = target.Nodename
	osRelease, err := target.OSRelease()
	if err != nil {
		log.Fatalf("could not retrieve OS release information: %v", err)
	}
	if strings.Contains(osRelease["ID_LIKE"], "suse") {
		joinConfiguration.NodeRegistration.KubeletExtraArgs["cni-bin-dir"] = "/usr/lib/cni"
	}
}

func createBootstrapToken(target string) string {
	client := kubernetes.GetAdminClientSet()

	internalCfg, err := kubeadmconfigutil.ConfigFileAndDefaultsToInternalConfig(caaspctl.KubeadmInitConfFile(), nil)
	if err != nil {
		log.Fatal("could not load init configuration")
	}

	bootstrapTokenRaw, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		log.Fatal("could not generate a new boostrap token")
	}

	bootstrapToken, err := kubeadmapi.NewBootstrapTokenString(bootstrapTokenRaw)
	if err != nil {
		log.Fatal("could not generate a new boostrap token")
	}

	internalCfg.BootstrapTokens = []kubeadmapi.BootstrapToken{
		kubeadmapi.BootstrapToken{
			Token:       bootstrapToken,
			Description: fmt.Sprintf("Bootstrap token for %s machine join", target),
			TTL: &metav1.Duration{
				Duration: 15 * time.Minute,
			},
			Usages: []string{"signing", "authentication"},
			Groups: []string{"system:bootstrappers:kubeadm:default-node-token"},
		},
	}

	if err := kubeadmtokenphase.CreateNewTokens(client, internalCfg.BootstrapTokens); err != nil {
		log.Fatal("could not create new bootstrap token")
	}

	return bootstrapTokenRaw
}

func joinConfigFileAndDefaultsToInternalConfig(cfgPath string) (*kubeadmapi.JoinConfiguration, error) {
	internalcfg := &kubeadmapi.JoinConfiguration{}

	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	if err := kubeadmconfigutil.DetectUnsupportedVersion(b); err != nil {
		return nil, err
	}

	gvkmap, err := kubeadmutil.SplitYAMLDocuments(b)
	if err != nil {
		return nil, err
	}

	joinBytes := []byte{}
	for gvk, bytes := range gvkmap {
		if gvk.Kind == constants.JoinConfigurationKind {
			joinBytes = bytes
			// verify the validity of the YAML
			strict.VerifyUnmarshalStrict(bytes, gvk)
		}
	}

	if len(joinBytes) == 0 {
		return nil, errors.New("invalid config")
	}

	if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), joinBytes, internalcfg); err != nil {
		return nil, err
	}

	return internalcfg, nil
}
