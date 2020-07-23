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

package skuba

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

const (
	ConfigMapName             = "skuba-config"
	SkubaConfigurationKeyName = "SkubaConfiguration"
)

type SkubaConfiguration struct {
	AddonsVersion kubernetes.AddonsVersion
}

func GetSkubaConfiguration(client clientset.Interface) (*SkubaConfiguration, error) {
	skubaConfiguration := &SkubaConfiguration{}
	configMap, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(ConfigMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return skubaConfiguration, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal([]byte(configMap.Data[SkubaConfigurationKeyName]), skubaConfiguration); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling SkubaConfiguration")
	}
	return skubaConfiguration, nil
}

func UpdateSkubaConfiguration(client clientset.Interface, skubaConfiguration *SkubaConfiguration) error {
	marshaledSkubaConfiguration, err := yaml.Marshal(skubaConfiguration)
	if err != nil {
		return errors.Wrap(err, "error marshaling SkubaConfiguration")
	}
	configMap := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigMapName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{SkubaConfigurationKeyName: string(marshaledSkubaConfiguration)},
	}
	if _, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(&configMap); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create configmap")
		}

		if _, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Update(&configMap); err != nil {
			return errors.Wrap(err, "unable to update configmap")
		}
	}
	return nil
}
