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
	// SecretName is secret name which stores dex client secret
	SecretName = "oidc-dex-client-secret"
)

// ClientSecret is the client secret used to
// authenticate auth client to auth server (dex)
type ClientSecret string

const (
	// Gangway client secret
	Gangway ClientSecret = "client-secret-gangway"
)

func (cs ClientSecret) String() string {
	return string(cs)
}

// CreateDexClientSecret creates client secret which is used by
// auth client (gangway) to authenticate to auth server (dex)
//
// Dex could generates more client secret at here if there are
// multiple auth clients
func CreateDexClientSecret() error {
	clientSecretGangway := make([]byte, 12)
	_, err := rand.Read(clientSecretGangway)
	if err != nil {
		return errors.Errorf("unable to generate client secret for gangway %v", err)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			Gangway.String(): clientSecretGangway,
		},
	}

	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "could not get admin client set")
	}
	if err := apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Wrap(err, "error when creating dex secret")
	}

	return nil
}

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
	certIPs := make([]net.IP, len(cfg.ClusterConfiguration.APIServer.CertSANs))
	for idx, san := range cfg.ClusterConfiguration.APIServer.CertSANs {
		certIPs[idx] = net.ParseIP(san)
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
