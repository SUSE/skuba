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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/pkg/errors"
)

// Bootstrap initializes the first master node of the cluster
//
// FIXME: being this a part of the go API accept the toplevel directory instead
//        of using the PWD
func Bootstrap(bootstrapConfiguration deployments.BootstrapConfiguration, target *deployments.Target) error {
	if clientSet, err := kubernetes.GetAdminClientSet(); err == nil {
		_, err := clientSet.ServerVersion()
		if err == nil {
			return errors.New("cluster is already bootstrapped")
		}
	}

	initConfiguration, err := LoadInitConfigurationFromFile(skuba.KubeadmInitConfFile())
	if err != nil {
		return errors.Wrapf(err, "could not parse %s file", skuba.KubeadmInitConfFile())
	}

	fmt.Println("[bootstrap] updating init configuration with target information")
	if err := addTargetInformationToInitConfiguration(target, initConfiguration); err != nil {
		return errors.Wrap(err, "unable to add target information to init configuration")
	}

	setHyperkubeImageToInitConfiguration(initConfiguration)
	setContainerImages(initConfiguration)
	setApiserverAdmissionPlugins(initConfiguration)
	finalInitConfigurationContents, err := kubeadmconfigutil.MarshalInitConfigurationToBytes(initConfiguration, schema.GroupVersion{
		Group:   "kubeadm.k8s.io",
		Version: "v1beta1",
	})

	if err != nil {
		return errors.Wrap(err, "could not marshal configuration")
	}

	fmt.Println("[bootstrap] writing init configuration for node")
	if err := ioutil.WriteFile(skuba.KubeadmInitConfFile(), finalInitConfigurationContents, 0600); err != nil {
		return errors.Wrap(err, "error writing init configuration")
	}

	fmt.Println("[bootstrap] applying init configuration to node")
	err = target.Apply(
		bootstrapConfiguration,
		"kubernetes.bootstrap.upload-secrets",
		"kernel.load-modules",
		"kernel.configure-parameters",
		"cri.start",
		"kubelet.configure",
		"kubelet.enable",
		"kubeadm.init",
		"psp.deploy",
		"kured.deploy",
		"skuba-update.start",
	)

	if err != nil {
		return err
	}

	err = downloadSecrets(target)
	if err != nil {
		return err
	}

	// deploy cni only after downloadSecrets because
	// we need to generate cilium etcd certs
	err = target.Apply(nil,
		"cni.deploy")
	if err != nil {
		return err
	}

	fmt.Printf("[bootstrap] successfully bootstrapped node %q with Kubernetes: %q\n", target.Target, kubernetes.LatestVersion().String())
	return nil
}

func downloadSecrets(target *deployments.Target) error {
	os.MkdirAll(filepath.Join("pki", "etcd"), 0700)

	fmt.Printf("[bootstrap] downloading secrets from bootstrapped node %q\n", target.Target)
	for _, secretLocation := range deployments.Secrets {
		secretData, err := target.DownloadFileContents(filepath.Join("/etc/kubernetes", secretLocation))
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(secretLocation, []byte(secretData), 0600); err != nil {
			return err
		}
	}

	return nil
}

func addTargetInformationToInitConfiguration(target *deployments.Target, initConfiguration *kubeadmapi.InitConfiguration) error {
	if initConfiguration.NodeRegistration.KubeletExtraArgs == nil {
		initConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
	}
	initConfiguration.NodeRegistration.Name = target.Nodename
	initConfiguration.NodeRegistration.CRISocket = skuba.CRISocket
	initConfiguration.NodeRegistration.KubeletExtraArgs["hostname-override"] = target.Nodename
	initConfiguration.NodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = images.GetGenericImage(skuba.ImageRepository, "pause", kubernetes.CurrentComponentVersion(kubernetes.Pause))
	isSUSE, err := target.IsSUSEOS()
	if err != nil {
		return err
	}
	if isSUSE {
		initConfiguration.NodeRegistration.KubeletExtraArgs["cni-bin-dir"] = skuba.SUSECNIDir
	}
	return nil
}

func setHyperkubeImageToInitConfiguration(initConfiguration *kubeadmapi.InitConfiguration) {
	initConfiguration.UseHyperKubeImage = true
}

func setContainerImages(initConfiguration *kubeadmapi.InitConfiguration) {
	initConfiguration.ImageRepository = skuba.ImageRepository
	initConfiguration.KubernetesVersion = kubernetes.CurrentComponentVersion(kubernetes.Hyperkube)
	initConfiguration.Etcd.Local = &kubeadmapi.LocalEtcd{
		ImageMeta: kubeadmapi.ImageMeta{
			ImageRepository: skuba.ImageRepository,
			ImageTag:        kubernetes.CurrentComponentVersion(kubernetes.Etcd),
		},
	}
	initConfiguration.DNS.ImageMeta = kubeadmapi.ImageMeta{
		ImageRepository: skuba.ImageRepository,
		ImageTag:        kubernetes.CurrentComponentVersion(kubernetes.CoreDNS),
	}
}

func setApiserverAdmissionPlugins(initConfiguration *kubeadmapi.InitConfiguration) {
	if initConfiguration.APIServer.ControlPlaneComponent.ExtraArgs == nil {
		initConfiguration.APIServer.ControlPlaneComponent.ExtraArgs = map[string]string{}
	}
	initConfiguration.APIServer.ControlPlaneComponent.ExtraArgs["enable-admission-plugins"] = "NodeRestriction,PodSecurityPolicy"
}
