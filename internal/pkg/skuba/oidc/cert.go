/*
 * Copyright (c) 2020 SUSE LLC.
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

package oidc

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/util"
	"github.com/SUSE/skuba/pkg/skuba"
)

const (
	// caCertAndKeyBaseFileName defines OIDC's CA certificate and key base file name
	caCertAndKeyBaseFileName = "oidc-ca"
	// caKeyFileName defines OIDC's CA key file name
	caKeyFileName = "oidc-ca.key"
	// CACertFileName defines OIDC's CA certificate file name
	CACertFileName = "oidc-ca.crt"

	// DexServerCertAndKeyBaseFileName defines OIDC's dex server certificate and key base file name
	DexServerCertAndKeyBaseFileName = "oidc-dex-server"
	// GangwayServerCertAndKeyBaseFileName defines OIDC's gangway server certificate and key base file name
	GangwayServerCertAndKeyBaseFileName = "oidc-gangway-server"
)

// IsCACertAndKeyExist returns the OIDC CA certificate and key exist or not
func IsCACertAndKeyExist() (bool, bool) {
	var certExist, keyExist bool

	cert := filepath.Join(skuba.PkiDir(), CACertFileName)
	f, err := os.Stat(cert)
	if !os.IsNotExist(err) && !f.IsDir() {
		certExist = true
	}

	key := filepath.Join(skuba.PkiDir(), caKeyFileName)
	f, err = os.Stat(key)
	if !os.IsNotExist(err) && !f.IsDir() {
		keyExist = true
	}

	return certExist, keyExist
}

// GenerateServerCSRAndKey generates server CSR and key and stores to local disk
func GenerateServerCSRAndKey(commonName string, sans []string, localBaseFileName string) error {
	csr, key, err := util.NewServerCSRAndKey(commonName, sans)
	if err != nil {
		return err
	}

	if err := pkiutil.WriteCSR(skuba.PkiDir(), localBaseFileName, csr); err != nil {
		return err
	}

	if err := pkiutil.WriteKey(skuba.PkiDir(), localBaseFileName, key); err != nil {
		return err
	}

	fmt.Printf("Generating %s server CSR and key to %s\n", commonName, filepath.Join(skuba.PkiDir(), localBaseFileName))

	return nil
}

// TryToUseLocalServerCert tries to load local OIDC server certificate if present
// and checks if the local OIDC server certificate is issued by custom OIDC CA certificate
func TryToUseLocalServerCert(client clientset.Interface, localServerCertBaseFileName, secretName string) error {
	serverCertFilePath := filepath.Join(skuba.PkiDir(), fmt.Sprintf("%s.crt", localServerCertBaseFileName))
	certFi, err := os.Stat(serverCertFilePath)
	if os.IsNotExist(err) || certFi.IsDir() {
		// The server cert do not exist
		return errors.New("the server cert do not exist")
	}

	// load local server cert/key pair
	cert, key, err := pkiutil.TryLoadCertAndKeyFromDisk(skuba.PkiDir(), localServerCertBaseFileName)
	if err != nil {
		return err
	}

	// load custom OIDC CA cert `oidc-ca.crt`
	caCert, err := pkiutil.TryLoadCertFromDisk(skuba.PkiDir(), caCertAndKeyBaseFileName)
	if err != nil {
		return err
	}

	// upload OIDC server cert to Secret
	if err := util.CreateOrUpdateCertToSecret(client, caCert, cert, key, secretName); err != nil {
		return err
	}

	return nil
}

// SignServerWithLocalCACertAndKey signs the OIDC server certificate by local OIDC CA cert/key pair if present
// otherwise, it signs the OIDC server certificate by cluster CA cert/key pair
func SignServerWithLocalCACertAndKey(client clientset.Interface, certCN, controlPlaneHost, certSecretName string) error {
	var caCert, cert *x509.Certificate
	var caKey, key crypto.Signer
	var err error

	// sign the server certificate by CA cert/key pair
	oidcCACertExist, oidcCAKeyExist := IsCACertAndKeyExist()
	if oidcCACertExist {
		if !oidcCAKeyExist {
			return errors.New("OIDC CA key not found")
		}

		// load custom OIDC CA cert/key pair
		caCert, caKey, err = pkiutil.TryLoadCertAndKeyFromDisk(skuba.PkiDir(), caCertAndKeyBaseFileName)
		if err != nil {
			return err
		}
	} else {
		// load cluster CA cert/key pair
		caCert, caKey, err = pkiutil.TryLoadCertAndKeyFromDisk(skuba.PkiDir(), constants.CACertAndKeyBaseName)
		if err != nil {
			return err
		}
	}

	// generate server certificate
	cert, key, err = util.NewServerCertAndKey(caCert, caKey, certCN, []string{controlPlaneHost})
	if err != nil {
		return err
	}

	// upload server cert to Secret
	if err := util.CreateOrUpdateCertToSecret(client, caCert, cert, key, certSecretName); err != nil {
		return err
	}

	return nil
}
