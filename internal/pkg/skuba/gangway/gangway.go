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

package gangway

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
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

	"github.com/SUSE/skuba/internal/pkg/skuba/dex"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	node "github.com/SUSE/skuba/pkg/skuba/actions/node/bootstrap"
)

const (
	certName   = "oidc-gangway-cert"
	secretName = "oidc-gangway-secret"

	clientSecret = "client-secret"
	sessionKey   = "session-key"
)

// CreateGangwaySecret generates session key and read client secret from dex
func CreateGangwaySecret() error {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return errors.Errorf("unable to generate session key %v", err)
	}

	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "could not get admin client set")
	}

	// Read client secret from dex secret
	dexSecret, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(dex.SecretName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "could not retrieve the dex secret")
	}
	cs, ok := dexSecret.Data[dex.Gangway.String()]
	if !ok {
		return errors.Wrap(err, "could not find dex client secret")
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			clientSecret: cs,
			sessionKey:   []byte(base64.URLEncoding.EncodeToString(key)),
		},
	}

	if err := apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Wrap(err, "error when creating gangway secret")
	}

	return nil
}

// CreateGangwayCert creates a signed certificate for gangway
// with kubernetes CA certificate and key
func CreateGangwayCert() error {
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

	// Generate gangway certificate
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, &certutil.Config{
		CommonName:   "oidc-gangway",
		Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
		AltNames: certutil.AltNames{
			DNSNames: cfg.ClusterConfiguration.APIServer.CertSANs,
			IPs:      certIPs,
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		return errors.Errorf("error when creating gangway certificate %v", err)
	}
	privateKey, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return errors.Errorf("gangway private key marshal failed %v", err)
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
		return errors.Errorf("error when creating gangway secret %v", err)
	}

	return nil
}

// GetGangwayImage returns gangway image registry
func GetGangwayImage() string {
	return images.GetGenericImage(skuba.ImageRepository, "gangway",
		kubernetes.CurrentAddonVersion(kubernetes.Gangway))
}
