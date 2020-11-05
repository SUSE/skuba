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
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"sigs.k8s.io/yaml"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/deployments/ssh/assets"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	stateMap["kubelet.rootcert.upload"] = kubeletUploadRootCert
	stateMap["kubelet.servercert.create-and-upload"] = kubeletCreateAndUploadServerCert
	stateMap["kubelet.configure"] = kubeletConfigure
	stateMap["kubelet.enable"] = kubeletEnable
}

func kubeletUploadRootCert(t *Target, data interface{}) error {
	// Upload root ca cert
	caCertPath := filepath.Join(skuba.PkiDir(), kubernetes.KubeletCACertName)
	f, err := os.Stat(caCertPath)
	if err != nil {
		return err
	}
	if err := t.target.UploadFile(caCertPath, filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletCACertName), f.Mode()); err != nil {
		return err
	}

	// Upload root ca key on control plane node only
	if *t.target.Role == deployments.MasterRole {
		caKeyPath := filepath.Join(skuba.PkiDir(), kubernetes.KubeletCAKeyName)
		f, err = os.Stat(caKeyPath)
		if err != nil {
			return err
		}
		if err := t.target.UploadFile(caKeyPath, filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletCAKeyName), f.Mode()); err != nil {
			return err
		}
	}

	return nil
}

func kubeletCreateAndUploadServerCert(t *Target, data interface{}) error {
	// Read kubelet root ca certificate and key
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk(skuba.PkiDir(), kubernetes.KubeletCACertAndKeyBaseName)
	if err != nil {
		return errors.Wrap(err, "failure loading kubelet CA certificate authority")
	}

	host := t.target.Nodename
	altNames := certutil.AltNames{}
	if ip := net.ParseIP(host); ip != nil {
		altNames.IPs = append(altNames.IPs, ip)
	} else {
		altNames.DNSNames = append(altNames.DNSNames, host)
	}

	// Create AltNames with defaults DNSNames/IPs
	stdout, _, err := t.silentSsh("hostname", "-I")
	if err != nil {
		return err
	}
	for _, addr := range strings.Split(stdout, " ") {
		if ip := net.ParseIP(addr); ip != nil {
			altNames.IPs = append(altNames.IPs, ip)
		}
	}

	alternateIPs := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	alternateDNS := []string{"localhost"}
	if ip := net.ParseIP(t.target.Target); ip != nil {
		alternateIPs = append(alternateIPs, ip)
	} else {
		alternateDNS = append(alternateDNS, t.target.Target)
	}

	altNames.IPs = append(altNames.IPs, alternateIPs...)
	altNames.DNSNames = append(altNames.DNSNames, alternateDNS...)

	cfg := &pkiutil.CertConfig{
		Config: certutil.Config{
			CommonName: host,
			AltNames:   altNames,
			Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
	}
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, cfg)
	if err != nil {
		return errors.Wrap(err, "couldn't generate kubelet server certificate")
	}

	// Save kubelet server certificate and key to local temporarily
	if err := pkiutil.WriteCertAndKey(skuba.PkiDir(), host, cert, key); err != nil {
		return errors.Wrapf(err, "failure while saving kubelet server %s certificate and key", host)
	}

	// Upload server certificate and key
	certPath, keyPath := pkiutil.PathsForCertAndKey(skuba.PkiDir(), host)
	f, err := os.Stat(certPath)
	if err != nil {
		return err
	}
	if err := t.target.UploadFile(certPath, filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletServerCertName), f.Mode()); err != nil {
		return err
	}
	f, err = os.Stat(keyPath)
	if err != nil {
		return err
	}
	if err := t.target.UploadFile(keyPath, filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletServerKeyName), f.Mode()); err != nil {
		return err
	}

	// Remove local temporarily kubelet server certificate and key
	if err := os.Remove(certPath); err != nil {
		return err
	}
	if err := os.Remove(keyPath); err != nil {
		return err
	}

	return nil
}

func kubeletConfigure(t *Target, data interface{}) error {
	isSUSE, err := t.target.IsSUSEOS()
	if err != nil {
		return err
	}
	if isSUSE {
		if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service", assets.KubeletService, 0644); err != nil {
			return err
		}
		if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService, 0644); err != nil {
			return err
		}
	} else {
		if err := t.UploadFileContents("/lib/systemd/system/kubelet.service", assets.KubeletService, 0644); err != nil {
			return err
		}
		if err := t.UploadFileContents("/etc/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService, 0644); err != nil {
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
	f, err := os.Stat(cloudConfigFilePath)
	if os.IsNotExist(err) {
		return err
	}
	cloudConfigRuntimeFilePath := filepath.Join(constants.KubernetesDir, cloudConfigFile)
	if err := t.target.UploadFile(cloudConfigFilePath, cloudConfigRuntimeFilePath, f.Mode()); err != nil {
		return err
	}
	return nil
}
