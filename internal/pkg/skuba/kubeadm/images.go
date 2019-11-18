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

package kubeadm

import (
	"k8s.io/apimachinery/pkg/util/version"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

func setContainerImagesWithClusterVersion(initConfiguration *kubeadmapi.InitConfiguration, clusterVersion *version.Version) error {
	initConfiguration.UseHyperKubeImage = true
	initConfiguration.ImageRepository = skuba.ImageRepository
	initConfiguration.KubernetesVersion = kubernetes.ComponentVersionForClusterVersion(kubernetes.Hyperkube, clusterVersion)
	initConfiguration.Etcd.Local = &kubeadmapi.LocalEtcd{
		ImageMeta: kubeadmapi.ImageMeta{
			ImageRepository: skuba.ImageRepository,
			ImageTag:        kubernetes.ComponentVersionForClusterVersion(kubernetes.Etcd, clusterVersion),
		},
	}
	initConfiguration.DNS.ImageMeta = kubeadmapi.ImageMeta{
		ImageRepository: skuba.ImageRepository,
		ImageTag:        kubernetes.ComponentVersionForClusterVersion(kubernetes.CoreDNS, clusterVersion),
	}
	return nil
}
