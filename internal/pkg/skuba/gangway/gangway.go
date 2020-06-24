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
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"sigs.k8s.io/yaml"

	"github.com/SUSE/skuba/internal/pkg/skuba/dex"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/node"
	"github.com/SUSE/skuba/internal/pkg/skuba/util"
)

const (
	certCommonName = "oidc-gangway"
	secretCertName = "oidc-gangway-cert"

	sessionKey    = "session-key"
	secretKeyName = "oidc-gangway-secret"

	configmapName = "oidc-gangway-config"
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
			Name:      secretKeyName,
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

// gClientSecret is a global variable to hold
// the generated client secret or existed client secret
var gClientSecret string

// GetClientSecret returns the client secret from global variable if present
// or from gangway ConfigMap if present
// otherwise generates a new client secret and stores it in gClientSecret
func GetClientSecret(client clientset.Interface) string {
	// global client secret existed, returns it
	if gClientSecret != "" {
		return gClientSecret
	}

	// global client secret not exist, read from ConfigMap
	clientSecret, err := getClientSecretFromConfigMap(client)
	if err != nil || clientSecret == "" {
		// generate a new client secret if read ConfigMap error
		// or client secret is not exist
		clientSecret = dex.GenerateClientSecret()
	}

	// update global client secret
	gClientSecret = clientSecret

	return gClientSecret
}

func getClientSecretFromConfigMap(client clientset.Interface) (string, error) {
	type config struct {
		ClientSecret string `yaml:"clientSecret"`
	}

	cm, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(configmapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}

	data, ok := cm.Data["gangway.yaml"]
	if !ok {
		return "", nil
	}

	c := config{}
	if err := yaml.Unmarshal([]byte(data), &c); err != nil {
		return "", err
	}
	return c.ClientSecret, nil
}

// CreateCert creates a signed certificate for gangway
// with kubernetes CA certificate and key
func CreateCert(client clientset.Interface, pkiPath, kubeadmInitConfPath string) error {
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
	if err := util.CreateOrUpdateCertToSecret(client, caCert, cert, key, secretCertName); err != nil {
		return errors.Wrap(err, "unable to create/update cert to secret")
	}

	return nil
}

func GangwaySecretExists(client clientset.Interface) (bool, error) {
	_, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(secretKeyName, metav1.GetOptions{})
	return kubernetes.DoesResourceExistWithError(err)
}

func GangwayCertExists(client clientset.Interface) (bool, error) {
	_, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(secretCertName, metav1.GetOptions{})
	return kubernetes.DoesResourceExistWithError(err)
}

func RestartPods(client clientset.Interface) error {
	listOptions := metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", certCommonName)}
	return client.CoreV1().Pods(metav1.NamespaceSystem).DeleteCollection(&metav1.DeleteOptions{}, listOptions)
}
