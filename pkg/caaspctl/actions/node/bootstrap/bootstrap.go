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
	"log"
	"os"
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
	"suse.com/caaspctl/internal/pkg/caaspctl/kubernetes"
	"suse.com/caaspctl/pkg/caaspctl"
)

// Bootstrap initializes the first master node of the cluster
//
// FIXME: being this a part of the go API accept the toplevel directory instead
//        of using the PWD
// FIXME: error handling with `github.com/pkg/errors`
func Bootstrap(target *deployments.Target) error {
	initConfiguration, err := configFileAndDefaultsToInternalConfig(caaspctl.KubeadmInitConfFile())
	if err != nil {
		return fmt.Errorf("Could not parse %s file: %v", caaspctl.KubeadmInitConfFile(), err)
	}
	addTargetInformationToInitConfiguration(target, initConfiguration)
	setHyperkubeImageToInitConfiguration(initConfiguration)
	setContainerImages(initConfiguration)
	finalInitConfigurationContents, err := kubeadmconfigutil.MarshalInitConfigurationToBytes(initConfiguration, schema.GroupVersion{
		Group:   "kubeadm.k8s.io",
		Version: "v1beta1",
	})
	if err != nil {
		return fmt.Errorf("Could not marshal configuration: %v", err)
	}

	if err := ioutil.WriteFile(caaspctl.KubeadmInitConfFile(), finalInitConfigurationContents, 0600); err != nil {
		return fmt.Errorf("Error writing init configuration: %v", err)
	}

	err = target.Apply(
		nil,
		"kubernetes.bootstrap.upload-secrets",
		"kernel.load-modules",
		"kernel.configure-parameters",
		"cri.start",
		"kubelet.configure",
		"kubelet.enable",
		"kubeadm.init",
		"cni.deploy",
	)
	if err != nil {
		return err
	}

	return downloadSecrets(target)
}

func downloadSecrets(target *deployments.Target) error {
	os.MkdirAll(path.Join("pki", "etcd"), 0700)

	for _, secretLocation := range deployments.Secrets {
		secretData, err := target.DownloadFileContents(path.Join("/etc/kubernetes", secretLocation))
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(secretLocation, []byte(secretData), 0600); err != nil {
			return err
		}
	}

	return nil
}

func addTargetInformationToInitConfiguration(target *deployments.Target, initConfiguration *kubeadmapi.InitConfiguration) {
	if initConfiguration.NodeRegistration.KubeletExtraArgs == nil {
		initConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
	}
	initConfiguration.NodeRegistration.Name = target.Nodename
	initConfiguration.NodeRegistration.CRISocket = caaspctl.CRISocket
	initConfiguration.NodeRegistration.KubeletExtraArgs["hostname-override"] = target.Nodename
	initConfiguration.NodeRegistration.KubeletExtraArgs["pod-infra-container-image"] = images.GetGenericImage(caaspctl.ImageRepository, "pause", kubernetes.CurrentComponentVersion(kubernetes.Pause))
	osRelease, err := target.OSRelease()
	if err != nil {
		log.Fatalf("could not retrieve OS release information: %v", err)
	}
	if strings.Contains(osRelease["ID_LIKE"], caaspctl.SUSEOSID) {
		initConfiguration.NodeRegistration.KubeletExtraArgs["cni-bin-dir"] = caaspctl.SUSECNIDir
	}
}

func setHyperkubeImageToInitConfiguration(initConfiguration *kubeadmapi.InitConfiguration) {
	initConfiguration.UseHyperKubeImage = true
}

func setContainerImages(initConfiguration *kubeadmapi.InitConfiguration) {
	initConfiguration.ImageRepository = caaspctl.ImageRepository
	initConfiguration.KubernetesVersion = kubernetes.CurrentVersion
	initConfiguration.Etcd.Local = &kubeadmapi.LocalEtcd{
		ImageMeta: kubeadmapi.ImageMeta{
			ImageRepository: caaspctl.ImageRepository,
			ImageTag:        kubernetes.CurrentComponentVersion(kubernetes.Etcd),
		},
	}
	initConfiguration.DNS.ImageMeta = kubeadmapi.ImageMeta{
		ImageRepository: caaspctl.ImageRepository,
		ImageTag:        kubernetes.CurrentComponentVersion(kubernetes.CoreDNS),
	}
}

func configFileAndDefaultsToInternalConfig(cfgPath string) (*kubeadmapi.InitConfiguration, error) {
	internalcfg := &kubeadmapi.InitConfiguration{}

	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	internalcfg, err = kubeadmconfigutil.BytesToInternalConfig(b)
	if err != nil {
		return nil, err
	}

	return internalcfg, nil
}
