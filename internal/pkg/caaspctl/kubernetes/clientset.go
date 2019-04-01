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

package kubernetes

import (
	"k8s.io/klog"

	clientset "k8s.io/client-go/kubernetes"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"github.com/SUSE/caaspctl/pkg/caaspctl"
)

func GetAdminClientSet() *clientset.Clientset {
	client, err := kubeconfigutil.ClientSetFromFile(caaspctl.KubeConfigAdminFile())
	if err != nil {
		klog.Fatal("could not load admin kubeconfig file")
	}
	return client
}
