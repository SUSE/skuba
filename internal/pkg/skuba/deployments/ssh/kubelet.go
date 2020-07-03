/*
 * Copyright (c) 2019,2020 SUSE LLC.
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

package ssh

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"sigs.k8s.io/yaml"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/deployments/ssh/assets"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	stateMap["kubelet.rootcert.upload"] = kubeletUploadRootCert
	stateMap["kubelet.configure"] = kubeletConfigure
	stateMap["kubelet.enable"] = kubeletEnable
}

func kubeletUploadRootCert(t *Target, data interface{}) error {
	// Upload root ca cert
	if err := t.target.UploadFile(filepath.Join(skuba.PkiDir(), kubernetes.KubeletCACertName), filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletCACertName)); err != nil {
		return err
	}
	// Upload root ca key on control plane node only
	if *t.target.Role == deployments.MasterRole {
		if err := t.target.UploadFile(filepath.Join(skuba.PkiDir(), kubernetes.KubeletCAKeyName), filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletCAKeyName)); err != nil {
			return err
		}
		if _, _, err := t.silentSsh("chmod", "0400", filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletCAKeyName)); err != nil {
			return err
		}
	}

	return nil
}

func kubeletConfigure(t *Target, data interface{}) error {
	isSUSE, err := t.target.IsSUSEOS()
	if err != nil {
		return err
	}
	if isSUSE {
		if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service", assets.KubeletService); err != nil {
			return err
		}
		if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService); err != nil {
			return err
		}
	} else {
		if err := t.UploadFileContents("/lib/systemd/system/kubelet.service", assets.KubeletService); err != nil {
			return err
		}
		if err := t.UploadFileContents("/etc/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService); err != nil {
			return err
		}
	}

	cloudProvider, err := getCloudProvider()
	if err != nil {
		return err
	}
	switch cloudProvider {
	case "azure", "openstack", "vsphere":
		if err := uploadCloudProvider(t, cloudProvider); err != nil {
			return err
		}
	}

	_, _, err = t.ssh("systemctl", "daemon-reload")
	return err
}

func kubeletEnable(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "enable", "kubelet")
	return err
}

func getCloudProvider() (string, error) {
	data, err := ioutil.ReadFile(skuba.KubeadmInitConfFile())
	if err != nil {
		return "", err
	}
	type config struct {
		NodeRegistration struct {
			KubeletExtraArgs struct {
				CloudProvider string `json:"cloud-provider"`
			} `json:"kubeletExtraArgs"`
		} `json:"nodeRegistration"`
	}
	c := config{}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return "", err
	}
	return c.NodeRegistration.KubeletExtraArgs.CloudProvider, nil
}

func uploadCloudProvider(t *Target, cloudProvider string) error {
	cloudConfigFile := cloudProvider + ".conf"
	cloudConfigFilePath := filepath.Join(skuba.CloudDir(), cloudProvider, cloudConfigFile)
	if _, err := os.Stat(cloudConfigFilePath); os.IsNotExist(err) {
		return err
	}
	cloudConfigRuntimeFilePath := filepath.Join(constants.KubernetesDir, cloudConfigFile)
	if err := t.target.UploadFile(cloudConfigFilePath, cloudConfigRuntimeFilePath); err != nil {
		return err
	}
	if _, _, err := t.ssh("chmod", "0400", cloudConfigRuntimeFilePath); err != nil {
		return err
	}
	return nil
}
