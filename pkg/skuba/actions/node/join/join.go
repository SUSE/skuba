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

package join

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmtokenphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/bootstraptoken/node"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"

	"github.com/SUSE/skuba/internal/pkg/skuba/cni"
	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/node"
	"github.com/SUSE/skuba/internal/pkg/skuba/replica"
	"github.com/SUSE/skuba/pkg/skuba"
)

// Join joins a new machine to the cluster. The role of the machine will be
// provided by the JoinConfiguration, and will target Target node
func Join(client clientset.Interface, joinConfiguration deployments.JoinConfiguration, target *deployments.Target) error {
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return err
	}

	_, err = target.InstallNodePattern(deployments.KubernetesBaseOSConfiguration{
		CurrentVersion: currentClusterVersion.String(),
	})
	if err != nil {
		return err
	}

	var criConfigure string
	if _, err := os.Stat(skuba.CriDockerDefaultsConfFile()); err == nil {
		criConfigure = "cri.configure"
	}

	_, err = client.CoreV1().Nodes().Get(target.Nodename, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("[join] failed to join the node with name %q since a node with the same name already exists in the cluster\n", target.Nodename)
		return err
	}

	statesToApply := []string{"kubeadm.reset"}

	if joinConfiguration.Role == deployments.MasterRole {
		statesToApply = append(statesToApply, "kubernetes.join.upload-secrets")
	}

	statesToApply = append(statesToApply,
		"kernel.load-modules",
		"kernel.configure-parameters",
		"apparmor.start",
		criConfigure,
		"cri.start",
		"kubelet.rootcert.upload",
		"kubelet.servercert.create-and-upload",
		"kubelet.configure",
		"kubelet.enable",
		"kubeadm.join",
		"skuba-update.start")

	fmt.Println("[join] applying states to new node")

	if err := target.Apply(joinConfiguration, statesToApply...); err != nil {
		fmt.Printf("[join] failed to apply join to node %s\n", err)
		return err
	}

	if joinConfiguration.Role == deployments.MasterRole {
		if err := cni.CiliumUpdateConfigMap(client); err != nil {
			return err
		}
	}

	replicaHelper, err := replica.NewHelper(client)
	if err != nil {
		return err
	}
	if err := replicaHelper.UpdateNodes(); err != nil {
		return err
	}

	fmt.Println("[join] node successfully joined the cluster")
	return nil
}

// ConfigPath returns the configuration path for a specific Target; if this file does
// not exist, it will be created out of the template file
func ConfigPath(client clientset.Interface, role deployments.Role, target *deployments.Target) (string, error) {
	configPath := skuba.MachineConfFile(target.Target)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = skuba.TemplatePathForRole(role)
	}

	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return "", errors.Wrap(err, "could not get current cluster version")
	}

	joinConfiguration, err := node.LoadJoinConfigurationFromFile(configPath)
	if err != nil {
		return "", errors.Wrap(err, "error parsing configuration")
	}
	if err := addFreshTokenToJoinConfiguration(client, target.Target, joinConfiguration); err != nil {
		return "", errors.Wrap(err, "error adding Token to join configuration")
	}
	if err := addTargetInformationToJoinConfiguration(target, role, joinConfiguration, currentClusterVersion); err != nil {
		return "", errors.Wrap(err, "error adding target information to join configuration")
	}
	finalJoinConfigurationContents, err := kubeadmutil.MarshalToYamlForCodecs(joinConfiguration, schema.GroupVersion{
		Group:   "kubeadm.k8s.io",
		Version: kubeadm.GetKubeadmApisVersion(currentClusterVersion),
	}, kubeadmscheme.Codecs)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal configuration")
	}

	if err := ioutil.WriteFile(skuba.MachineConfFile(target.Target), finalJoinConfigurationContents, 0600); err != nil {
		return "", errors.Wrap(err, "error writing specific machine configuration")
	}

	return skuba.MachineConfFile(target.Target), nil
}

func addFreshTokenToJoinConfiguration(client clientset.Interface, target string, joinConfiguration *kubeadmapi.JoinConfiguration) error {
	if joinConfiguration.Discovery.BootstrapToken == nil {
		joinConfiguration.Discovery.BootstrapToken = &kubeadmapi.BootstrapTokenDiscovery{}
	}
	var err error
	joinConfiguration.Discovery.BootstrapToken.Token, err = createBootstrapToken(client, target)
	joinConfiguration.Discovery.TLSBootstrapToken = ""
	return err
}

func addTargetInformationToJoinConfiguration(target *deployments.Target, role deployments.Role, joinConfiguration *kubeadmapi.JoinConfiguration, clusterVersion *version.Version) error {
	if joinConfiguration.NodeRegistration.KubeletExtraArgs == nil {
		joinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
	}
	joinConfiguration.NodeRegistration.Name = target.Nodename
	joinConfiguration.NodeRegistration.CRISocket = skuba.CRISocket
	joinConfiguration.NodeRegistration.KubeletExtraArgs["hostname-override"] = target.Nodename
	joinConfiguration.NodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = kubernetes.ComponentContainerImageForClusterVersion(kubernetes.Pause, clusterVersion)
	isSUSE, err := target.IsSUSEOS()
	if err != nil {
		return errors.Wrap(err, "unable to get os info")
	}
	if isSUSE {
		joinConfiguration.NodeRegistration.KubeletExtraArgs["cni-bin-dir"] = skuba.SUSECNIDir
	}
	return nil
}

func createBootstrapToken(client clientset.Interface, target string) (string, error) {
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
