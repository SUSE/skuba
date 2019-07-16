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
	"fmt"

	"github.com/pkg/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/util"
	"github.com/SUSE/skuba/pkg/skuba"
	node "github.com/SUSE/skuba/pkg/skuba/actions/node/bootstrap"
)

const (
	imageName = "caasp-dex"

	certCommonName = "oidc-dex"
	secretName     = "oidc-dex-cert"
)

// CreateCert creates a signed certificate for dex
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

	// Generate dex certificate
	cert, key, err := util.NewServerCertAndKey(caCert, caKey,
		certCommonName, cfg.ClusterConfiguration.APIServer.CertSANs)
	if err != nil {
		return errors.Wrap(err, "could not genenerate dex server cert")
	}

	// Create or update secret resource
	if err := util.CreateOrUpdateCertToSecret(client, caCert, cert, key, certCommonName); err != nil {
		return errors.Wrap(err, "unable to create/update cert to secret")
	}

	return nil
}

// GetDexImage returns dex image registry
func GetDexImage() string {
	return images.GetGenericImage(skuba.ImageRepository, imageName,
		kubernetes.CurrentAddonVersion(kubernetes.Dex))
}

// GenerateClientSecret returns client secret which is used by
// auth client (gangway) to authenticate to auth server (dex)
//
// Due to this issue https://github.com/dexidp/dex/issues/1099
// client secret is not configurable through environment variable
// so, replace client secret in configmap by rendering
func GenerateClientSecret() string {
	b := make([]byte, 12)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
