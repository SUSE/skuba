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

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	pkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

const (
	gangwaySecretName = "oidc-gangway-secret"
	gangwayCertName   = "oidc-gangway-cert"

	gangwayClientSecret = "clientsecret"
	gangwaySessionKey   = "sessionkey"
)

var (
	gangwayCertConfg = certutil.Config{
		CommonName:   "oidc-gangway",
		Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
)

// CreateGangwaySecret generates client secret and session key
func CreateGangwaySecret() error {
	clientSecret := make([]byte, 12)
	_, err := rand.Read(clientSecret)
	if err != nil {
		return errors.Errorf("unable to generate client secret %v", err)
	}

	sessionKey := make([]byte, 32)
	_, err = rand.Read(sessionKey)
	if err != nil {
		return errors.Errorf("unable to generate session key %v", err)
	}

	gangwaySecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gangwaySecretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			gangwayClientSecret: clientSecret,
			gangwaySessionKey:   []byte(base64.URLEncoding.EncodeToString(sessionKey)),
		},
	}

	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "could not get admin client set")
	}
	if err := apiclient.CreateOrUpdateSecret(client, gangwaySecret); err != nil {
		return errors.Wrap(err, "error when creating gangway secret")
	}

	return nil
}

// CreateGangwayCert creates a signed certificate for gangway
// with kubernetes CA certificate and key
func CreateGangwayCert() error {
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk("pki", "ca")
	if err != nil {
		return errors.Errorf("unable to load kubernetes CA certificate and key %v", err)
	}
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, &gangwayCertConfg)
	if err != nil {
		return errors.Errorf("error when creating gangway certificate %v", err)
	}
	privateKey, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return errors.Errorf("gangway private key marshal failed %v", err)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gangwayCertName,
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
