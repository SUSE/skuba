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

package ssh

import (
	"crypto/x509"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

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

	cfg := &certutil.Config{
		CommonName: host,
		AltNames:   altNames,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
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
	if err := t.target.UploadFile(certPath, filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletServerCertName)); err != nil {
		return err
	}
	if err := t.target.UploadFile(keyPath, filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletServerKeyName)); err != nil {
		return err
	}
	if _, _, err := t.silentSsh("chmod", "0400", filepath.Join(kubernetes.KubeletCertAndKeyDir, kubernetes.KubeletServerKeyName)); err != nil {
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
		if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service", assets.KubeletService); err != nil {
			return err
		}
		if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService); err != nil {
			return err
		}
		if err := t.UploadFileContents("/etc/sysconfig/kubelet", assets.KubeletSysconfig); err != nil {
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

	if _, err := os.Stat(skuba.OpenstackCloudConfFile()); err == nil {
		if err := t.target.UploadFile(skuba.OpenstackCloudConfFile(), skuba.OpenstackConfigRuntimeFile()); err != nil {
			return err
		}
		if _, _, err = t.ssh("chmod", "0400", skuba.OpenstackConfigRuntimeFile()); err != nil {
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
