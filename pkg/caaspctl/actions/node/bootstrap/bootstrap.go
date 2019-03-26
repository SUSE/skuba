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
	"net"
	"os"
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
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

	/* deploy cni only after downloadSecrets because
	we need to generate cilium etcd certs */
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
	if ip := net.ParseIP(target.Target); ip != nil {
		initConfiguration.NodeRegistration.KubeletExtraArgs["node-ip"] = target.Target
	}
	initConfiguration.NodeRegistration.Name = target.Nodename
	initConfiguration.NodeRegistration.KubeletExtraArgs["hostname-override"] = target.Nodename
	osRelease, err := target.OSRelease()
	if err != nil {
		log.Fatalf("could not retrieve OS release information: %v", err)
	}
	if strings.Contains(osRelease["ID_LIKE"], "suse") {
		initConfiguration.NodeRegistration.KubeletExtraArgs["cni-bin-dir"] = "/usr/lib/cni"
	}
}

func setHyperkubeImageToInitConfiguration(initConfiguration *kubeadmapi.InitConfiguration) {
	initConfiguration.UseHyperKubeImage = true
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
