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

package metricsserver

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/util"
)

const (
	certCommonName = "metrics-server.kube-system.svc" // DO NOT CHANGE THE CN
	secretCertName = "metrics-server-cert"
)

// CreateCert creates a signed certificate for metrics-server
// with kubernetes CA certificate and key
func CreateCert(client clientset.Interface, pkiPath string) error {
	// Load kubernetes CA
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk(pkiPath, constants.CACertAndKeyBaseName)
	if err != nil {
		return errors.Wrap(err, "unable to load kubernetes CA certificate and key")
	}

	// Generate metrics-server certificate
	cert, key, err := util.NewServerCertAndKey(caCert, caKey, certCommonName, []string{})
	if err != nil {
		return errors.Wrap(err, "could not genenerate metrics-server server cert")
	}

	// Create or update certificate to secret
	if err := util.CreateOrUpdateCertToSecret(client, caCert, cert, key, secretCertName); err != nil {
		return errors.Wrap(err, "unable to create/update cert to secret")
	}

	return nil
}

// IsCertExist check the metrics-server certificate secret resource exist
func IsCertExist(client clientset.Interface) (bool, error) {
	_, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(secretCertName, metav1.GetOptions{})
	return kubernetes.DoesResourceExistWithError(err)
}
