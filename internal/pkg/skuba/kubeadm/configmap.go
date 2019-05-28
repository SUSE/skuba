/*
 * Copyright (c) 2019 SUSE LLC. All rights reserved.
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

package kubeadm

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	configutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/pkg/errors"
)

// GetAPIEndpointsFromConfigMap returns the api endpoint held in the config map
func GetAPIEndpointsFromConfigMap() ([]string, error) {
	apiEndpoints := []string{}
	clientSet, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting client set")
	}
	kubeadmConfig, err := clientSet.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(kubeadmconstants.KubeadmConfigConfigMap, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve the kubeadm-config configmap to get apiEndpoints")
	}
	clusterStatus, err := configutil.UnmarshalClusterStatus(kubeadmConfig.Data)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal cluster status from kubeadm-config configmap")
	}

	for node := range clusterStatus.APIEndpoints {
		apiEndpoints = append(apiEndpoints, clusterStatus.APIEndpoints[node].AdvertiseAddress)
	}

	return apiEndpoints, nil
}

// RemoveAPIEndpointFromConfigMap removes api endpoints from the config map
func RemoveAPIEndpointFromConfigMap(node *v1.Node) error {
	clientSet, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "Error getting client set")
	}
	kubeadmConfig, err := clientSet.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(kubeadmconstants.KubeadmConfigConfigMap, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "could not retrieve the kubeadm-config configmap to change the apiEndpoints")
	}
	clusterStatus := &kubeadmapi.ClusterStatus{}
	if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), []byte(kubeadmConfig.Data[kubeadmconstants.ClusterStatusConfigMapKey]), clusterStatus); err != nil {
		return errors.Wrap(err, "could not unmarshal cluster status from kubeadm-config configmap")
	}
	delete(clusterStatus.APIEndpoints, node.ObjectMeta.Name)
	clusterStatusYaml, err := configutil.MarshalKubeadmConfigObject(clusterStatus)
	if err != nil {
		return errors.Wrap(err, "could not marshal modified cluster status")
	}
	clientSet, err = kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "Error getting client set")
	}
	_, err = clientSet.CoreV1().ConfigMaps(metav1.NamespaceSystem).Update(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeadmconstants.KubeadmConfigConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			kubeadmconstants.ClusterConfigurationConfigMapKey: kubeadmConfig.Data[kubeadmconstants.ClusterConfigurationConfigMapKey],
			kubeadmconstants.ClusterStatusConfigMapKey:        string(clusterStatusYaml),
		},
	})
	if err != nil {
		return errors.Wrap(err, "could not update kubeadm-config configmap")
	}
	return nil
}
