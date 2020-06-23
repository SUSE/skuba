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
	"context"
	"encoding/base64"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

// CreateOrUpdateToSecret create/update key=value to secret
func CreateOrUpdateToSecret(client clientset.Interface, secretName, key string, value []byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			key: []byte(base64.URLEncoding.EncodeToString(value)),
		},
	}

	if err := apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Wrapf(err, "error when create/update secret %s", secretName)
	}

	return nil
}

// IsSecretExist checks if the secret in namespace kube-system exist
func IsSecretExist(client clientset.Interface, secretName string) (bool, error) {
	_, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(context.TODO(), secretName, metav1.GetOptions{})
	return kubernetes.DoesResourceExistWithError(err)
}
