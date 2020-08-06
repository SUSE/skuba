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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

func certSANsToAltNAmes(sans []string) certutil.AltNames {
	altNames := certutil.AltNames{}
	for _, san := range sans {
		if ip := net.ParseIP(san); ip != nil {
			altNames.IPs = append(altNames.IPs, ip)
		} else {
			altNames.DNSNames = append(altNames.DNSNames, san)
		}
	}
	return altNames
}

// NewServerCertAndKey creates new CSR and key by
// passing the server common name and server SANs
func NewServerCSRAndKey(commonName string, sans []string) (*x509.CertificateRequest, crypto.Signer, error) {
	cfg := &pkiutil.CertConfig{
		Config: certutil.Config{
			CommonName:   commonName,
			Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
			AltNames:     certSANsToAltNAmes(sans),
			Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
	}
	return pkiutil.NewCSRAndKey(cfg)
}

// NewServerCertAndKey creates new certificate and key by
// passing the certificate authority certificate and key
// and server common name and server SANs
func NewServerCertAndKey(
	caCert *x509.Certificate, caKey crypto.Signer,
	commonName string, sans []string,
) (*x509.Certificate, crypto.Signer, error) {
	if caCert == nil {
		return nil, nil, errors.New("invalid input")
	}

	cfg := &pkiutil.CertConfig{
		Config: certutil.Config{
			CommonName:   commonName,
			Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
			AltNames:     certSANsToAltNAmes(sans),
			Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
	}
	return pkiutil.NewCertAndKey(caCert, caKey, cfg)
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
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
			Labels:    map[string]string{"caasp.suse.com/skuba-addon": "true"},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       pkiutil.EncodeCertPEM(cert),
			corev1.TLSPrivateKeyKey: privateKey,
			"ca.crt":                pkiutil.EncodeCertPEM(caCert),
		},
	}

	if err = apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Errorf("error when create/update secret %v", err)
	}

	return nil
}
