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

package certutil

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"net"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

const (
	// CACertKey is the key for tls CA certificate in a TLS secert.
	CACertKey = "ca.crt"
)

// CreateOrUpdateServerCertAndKeyToSecret creates/updates a signed certificate
// with provided CA certificate and key path, and SAN.
//
// User might replace server certificate and key manually.
// Therefore, check existing server server certificate is signed by the ca cert.
//   If not, skip signed the server certificate and key, and return.
//   Otherwise, signed the server certificate and key, and create/update
//   server certificate and key to secret resource.
func CreateOrUpdateServerCertAndKeyToSecret(
	client clientset.Interface,
	caCert *x509.Certificate, caKey crypto.Signer,
	commonName string, certSANs []string,
	secretName string,
) (bool, error) {
	secret, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(secretName, metav1.GetOptions{})
	serverCertExist, err := kubernetes.DoesResourceExistWithError(err)
	if err != nil {
		return false, errors.Wrap(err, "unable to determine if server certificate exists")
	}

	// Check existing server certificate is signed by the ca cert
	if serverCertExist {
		serverCertPEM, ok := secret.Data[corev1.TLSCertKey]
		if !ok {
			return false, errors.Wrapf(err, "server certificate %s not exists", corev1.TLSCertKey)
		}
		serverCertBlock, _ := pem.Decode(serverCertPEM)
		if serverCertBlock == nil {
			return false, errors.New("failed to decode server certificate PEM")
		}
		serverCert, err := x509.ParseCertificate(serverCertBlock.Bytes)
		if err != nil {
			return false, errors.Wrap(err, "failed to parse server certificate block")
		}

		// Verify server certificate is signed by CA certificate
		root := x509.NewCertPool()
		root.AddCert(caCert)
		if _, err := serverCert.Verify(x509.VerifyOptions{Roots: root}); err != nil {
			klog.Info("server certificate is not signed by ca certificate")
			return false, nil
		}
	}

	// Generate server certificate and key
	cert, key, err := NewServerCertAndKey(caCert, caKey, commonName, certSANs)
	if err != nil {
		return false, errors.Wrap(err, "could not new server cert and key")
	}

	// Create or update certificate and key to secret
	if err := CreateOrUpdateCertAndKeyToSecret(client, caCert, cert, key, secretName); err != nil {
		return false, errors.Wrap(err, "unable to create/update cert and key to secret")
	}

	if serverCertExist {
		return true, nil
	}
	return false, nil
}

// NewServerCertAndKey creates new certificate and key by
// passing the certificate authority certificate and key
// and server common name and server SANs
func NewServerCertAndKey(
	caCert *x509.Certificate, caKey crypto.Signer,
	commonName string, certSANs []string,
) (*x509.Certificate, crypto.Signer, error) {
	if caCert == nil {
		return nil, nil, errors.New("invalid input")
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
		return nil, nil, errors.Wrap(err, "new cert and key")
	}

	return cert, key, nil
}

// CreateOrUpdateCertAndKeyToSecret creates or update
// certificate and key to secret resource
func CreateOrUpdateCertAndKeyToSecret(
	client clientset.Interface,
	caCert *x509.Certificate,
	cert *x509.Certificate, key crypto.Signer,
	secretName string,
) error {
	if caCert == nil || cert == nil {
		return errors.New("invalid input")
	}

	privateKey, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return errors.Wrap(err, "private key marshal to pem failed")
	}

	// Write certificate into secret resource
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			CACertKey:               pkiutil.EncodeCertPEM(caCert),
			corev1.TLSCertKey:       pkiutil.EncodeCertPEM(cert),
			corev1.TLSPrivateKeyKey: privateKey,
		},
	}

	if err = apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Wrap(err, "create/update secret")
	}

	return nil
}
