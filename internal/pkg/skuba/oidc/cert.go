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
	"github.com/pkg/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/util"
)

// CreateServerCert creates a server certificate with subject alternative name (SAN)
// issued by kubernetes CA cert/key pair
func CreateServerCert(client clientset.Interface, pkiPath, certCN, controlPlaneHost, certSecretName string) error {
	// load kubernetes CA
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk(pkiPath, constants.CACertAndKeyBaseName)
	if err != nil {
		return errors.Errorf("unable to load kubernetes CA certificate and key %v", err)
	}

	// generate server certificate
	cert, key, err := util.NewServerCertAndKey(caCert, caKey, certCN, []string{controlPlaneHost})
	if err != nil {
		return errors.Wrapf(err, "could not genenerate %s server cert", certCN)
	}

	// create or update secret resource
	if err := util.CreateOrUpdateCertToSecret(client, caCert, cert, key, certSecretName); err != nil {
		return errors.Wrapf(err, "unable to create/update %s cert to secret %s", certCN, certSecretName)
	}

	return nil
}
