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
	"encoding/base64"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/util"
	"github.com/SUSE/skuba/pkg/skuba"
	node "github.com/SUSE/skuba/pkg/skuba/actions/node/bootstrap"
)

const (
	imageName = "gangway"

	certCommonName = "oidc-gangway-cert"
	secretName     = "oidc-gangway-secret"

	sessionKey = "session-key"
)

// GenerateSessionKey generates session key
func GenerateSessionKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, errors.Errorf("unable to generate session key %v", err)
	}

	return key, nil
}

// CreateOrUpdateSessionKeyToSecret create/update session key to secret
func CreateOrUpdateSessionKeyToSecret(client clientset.Interface, key []byte) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			sessionKey: []byte(base64.URLEncoding.EncodeToString(key)),
		},
	}

	if err := apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Wrap(err, "error when create/update session key to secret resource")
	}

	return nil
}

// CreateCert creates a signed certificate for gangway
// with kubernetes CA certificate and key
func CreateCert(
	client clientset.Interface,
	pkiPath, kubeadmInitConfPath string,
) error {
	// Load kubernetes CA
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk(pkiPath, constants.CACertAndKeyBaseName)
	if err != nil {
		return errors.Errorf("unable to load kubernetes CA certificate and key %v", err)
	}

	// Load kubeadm-init.conf to get certificate SANs
	cfg, err := node.LoadInitConfigurationFromFile(kubeadmInitConfPath)
	if err != nil {
		return errors.Wrapf(err, "could not parse %s file", kubeadmInitConfPath)
	}

	// Generate gangway certificate
	cert, key, err := util.NewServerCertAndKey(caCert, caKey,
		certCommonName, cfg.ClusterConfiguration.APIServer.CertSANs)
	if err != nil {
		return errors.Wrap(err, "could not genenerate gangway server cert")
	}

	// Create or update certificate to secret
	if err := util.CreateOrUpdateCertToSecret(client, caCert, cert, key, secretName); err != nil {
		return errors.Wrap(err, "unable to create/update cert to secret")
	}

	return nil
}

// GetGangwayImage returns gangway image registry
func GetGangwayImage() string {
	return images.GetGenericImage(skuba.ImageRepository, imageName,
		kubernetes.CurrentAddonVersion(kubernetes.Gangway))
}
