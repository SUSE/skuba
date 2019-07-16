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

package util

import (
	"crypto"
	"crypto/x509"
	"net"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

// NewServerCertAndKey creates new certificate and key by
// passing the certificate authority certificate and key
// and server common name and server SANs
func NewServerCertAndKey(
	caCert *x509.Certificate, caKey crypto.Signer,
	commonName string, certSANs []string,
) (*x509.Certificate, crypto.Signer, error) {
	if caCert == nil {
		return nil, nil, errors.Errorf("invalid input")
	}

	certDNSNames := make([]string, 0)
	certIPs := make([]net.IP, 0)
	for _, san := range certSANs {
		dnsOrIP := net.ParseIP(san)
		if dnsOrIP != nil {
			certIPs = append(certIPs, dnsOrIP)
		} else {
			certDNSNames = append(certDNSNames, san)
		}
	}

	// Generate certificate
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, &certutil.Config{
		CommonName:   commonName,
		Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
		AltNames: certutil.AltNames{
			DNSNames: certDNSNames,
			IPs:      certIPs,
		},
		Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		return nil, nil, errors.Errorf("error when creating certificate %v", err)
	}

	return cert, key, nil
}

// CreateOrUpdateCertToSecret creates or update
// certificate to secret resource
func CreateOrUpdateCertToSecret(
	client clientset.Interface,
	caCert *x509.Certificate,
	cert *x509.Certificate, key crypto.Signer,
	secretName string,
) error {
	if caCert == nil || cert == nil {
		return errors.Errorf("invalid input")
	}

	privateKey, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return errors.Errorf("private key marshal failed %v", err)
	}

	// Write certificate into secret resource
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
		},
		Type: v1.SecretTypeTLS,
		Data: map[string][]byte{
			v1.TLSCertKey:              pkiutil.EncodeCertPEM(cert),
			v1.TLSPrivateKeyKey:        privateKey,
			v1.ServiceAccountRootCAKey: pkiutil.EncodeCertPEM(caCert),
		},
	}

	if err = apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Errorf("error when create/update secret %v", err)
	}

	return nil
}
