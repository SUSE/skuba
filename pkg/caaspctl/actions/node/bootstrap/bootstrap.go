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
	"path"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/kubernetes"
	"github.com/SUSE/caaspctl/pkg/caaspctl"
	"github.com/pkg/errors"
)

// Bootstrap initializes the first master node of the cluster
//
// FIXME: being this a part of the go API accept the toplevel directory instead
//        of using the PWD
func Bootstrap(bootstrapConfiguration deployments.BootstrapConfiguration, target *deployments.Target) error {
	initConfiguration, err := LoadInitConfigurationFromFile(caaspctl.KubeadmInitConfFile())
	if err != nil {
		return errors.Wrapf(err, "could not parse %s file", caaspctl.KubeadmInitConfFile())
	}
	fmt.Println("[bootstrap] updating init configuration with target information")
	addTargetInformationToInitConfiguration(target, initConfiguration)
	setHyperkubeImageToInitConfiguration(initConfiguration)
	setContainerImages(initConfiguration)
	// FIXME: this is required when cilium uses the cluster etcd and we have floating IP addresses, remove me after
	// cilium uses its own etcd cluster
	setEtcdExtraSANs(target, initConfiguration)
	finalInitConfigurationContents, err := kubeadmconfigutil.MarshalInitConfigurationToBytes(initConfiguration, schema.GroupVersion{
		Group:   "kubeadm.k8s.io",
		Version: "v1beta1",
	})
	if err != nil {
		return errors.Wrap(err, "could not marshal configuration")
	}

	fmt.Println("[bootstrap] writing init configuration for node")
	if err := ioutil.WriteFile(caaspctl.KubeadmInitConfFile(), finalInitConfigurationContents, 0600); err != nil {
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
		"cni.render",
		"cni.deploy")
	if err != nil {
		return err
	}

	return nil
}

func downloadSecrets(target *deployments.Target) error {
	os.MkdirAll(path.Join("pki", "etcd"), 0700)

	for _, secretLocation := range deployments.Secrets {
		secretData, err := target.DownloadFileContents(path.Join("/etc/kubernetes", secretLocation))
		if err != nil {
			return err
		}
		fmt.Println("[bootstrap] downloading secrets from new node")
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
	isSUSE, err := target.IsSUSEOS()
	if err != nil {
		klog.Fatal(err)
	}
	if isSUSE {
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

func setEtcdExtraSANs(target *deployments.Target, initConfiguration *kubeadmapi.InitConfiguration) {
	if initConfiguration.Etcd.Local.ServerCertSANs == nil {
		initConfiguration.Etcd.Local.ServerCertSANs = []string{}
	}
	initConfiguration.Etcd.Local.ServerCertSANs = append(initConfiguration.Etcd.Local.ServerCertSANs, target.Target)
}
