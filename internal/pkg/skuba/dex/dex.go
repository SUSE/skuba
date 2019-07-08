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

package dex

import (
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	node "github.com/SUSE/skuba/pkg/skuba/actions/node/bootstrap"
)

const (
	certName = "oidc-dex-cert"
)

// CreateDexCert creates a signed certificate for dex
// with kubernetes CA certificate and key
func CreateDexCert() error {
	// Load kubernetes CA
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk("pki", "ca")
	if err != nil {
		return errors.Errorf("unable to load kubernetes CA certificate and key %v", err)
	}

	// Load kubeadm-init.conf to get certificate SANs
	cfg, err := node.LoadInitConfigurationFromFile(skuba.KubeadmInitConfFile())
	if err != nil {
		return errors.Wrapf(err, "could not parse %s file", skuba.KubeadmInitConfFile())
	}
	certIPs := make([]net.IP, 0)
	for _, san := range cfg.ClusterConfiguration.APIServer.CertSANs {
		if ip := net.ParseIP(san); ip != nil {
			certIPs = append(certIPs, ip)
		}
	}

	// Generate dex certificate
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, &certutil.Config{
		CommonName:   "oidc-dex",
		Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
		AltNames: certutil.AltNames{
			DNSNames: cfg.ClusterConfiguration.APIServer.CertSANs,
			IPs:      certIPs,
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		return errors.Errorf("error when creating dex certificate %v", err)
	}
	privateKey, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return errors.Errorf("dex private key marshal failed %v", err)
	}

	// Write certificate into secret resource
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: metav1.NamespaceSystem,
		},
		Type: v1.SecretTypeTLS,
		Data: map[string][]byte{
			v1.TLSCertKey:              pkiutil.EncodeCertPEM(cert),
			v1.TLSPrivateKeyKey:        privateKey,
			v1.ServiceAccountRootCAKey: pkiutil.EncodeCertPEM(caCert),
		},
	}

	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "unable to get admin client set")
	}
	if err = apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Errorf("error when creating dex secret %v", err)
	}

	return nil
}

// GetDexImage returns dex image registry
func GetDexImage() string {
	return images.GetGenericImage(skuba.ImageRepository, "caasp-dex",
		kubernetes.CurrentAddonVersion(kubernetes.Dex))
}

// GetClientSecretGangway returns client secret which is used by
// auth client (gangway) to authenticate to auth server (dex)
//
// Due to this issue https://github.com/dexidp/dex/issues/1099
// client secret is not configurable through environment variable
// so, replace client secret in configmap by rendering
func GetClientSecretGangway() string {
	b := make([]byte, 12)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
