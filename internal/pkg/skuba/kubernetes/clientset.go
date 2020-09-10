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

package kubernetes

import (
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/SUSE/skuba/pkg/skuba"
)

func GetAdminClientSetWithConfig() (clientset.Interface, *rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", skuba.KubeConfigAdminFile())
	if err != nil {
		return nil, nil, err
	}
	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return client, config, nil
}

func GetAdminClientSet() (clientset.Interface, error) {
	client, _, err := GetAdminClientSetWithConfig()
	return client, err
}
