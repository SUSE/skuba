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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
)

const (
	// CertCommonName is the gangway server certificate CN
	CertCommonName = "oidc-gangway"
	// PodLabelName is the gangway pod label name
	PodLabelName = "app=oidc-gangway"
	// CertSecretName is the gangway certificate secret name
	CertSecretName = "oidc-gangway-cert"

	sessionKey = "session-key"
	// SessionKeySecretName is the gangway session key secret name
	SessionKeySecretName = "oidc-gangway-secret"
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
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SessionKeySecretName,
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
