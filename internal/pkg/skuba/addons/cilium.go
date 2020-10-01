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

package addons

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"

	"github.com/SUSE/skuba/internal/pkg/skuba/addons/cilium_manifests"
	"github.com/SUSE/skuba/internal/pkg/skuba/cni"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	registerAddon(kubernetes.Cilium, CniAddOn, renderCiliumTemplate, renderCiliumPreflightTemplate, ciliumCallbacks{}, normalPriority, []getImageCallback{GetCiliumOperatorImage, GetCiliumImage})
}

func GetCiliumOperatorImage(clusterVersion *version.Version, imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository(clusterVersion), "cilium-operator", imageTag)
}

func GetCiliumImage(clusterVersion *version.Version, imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository(clusterVersion), "cilium", imageTag)
}

func (renderContext renderContext) CiliumOperatorImage() string {
	return GetCiliumOperatorImage(renderContext.config.ClusterVersion, kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, renderContext.config.ClusterVersion).Version)
}

func (renderContext renderContext) CiliumImage() string {
	return GetCiliumImage(renderContext.config.ClusterVersion, kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, renderContext.config.ClusterVersion).Version)
}

func renderCiliumTemplate(addonConfiguration AddonConfiguration) string {
	ciliumVersion := kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, addonConfiguration.ClusterVersion).Version
	switch {
	case strings.HasPrefix(ciliumVersion, "1.5"):
		return cilium_manifests.Manifestv15
	case strings.HasPrefix(ciliumVersion, "1.6"):
		return cilium_manifests.Manifestv16
	case strings.HasPrefix(ciliumVersion, "1.7"):
		return cilium_manifests.Manifestv17
	}
	panic(fmt.Sprintf("invalid cilium addon version: %s", ciliumVersion))
}

func renderCiliumPreflightTemplate(addonConfiguration AddonConfiguration) string {
	ciliumVersion := kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, addonConfiguration.ClusterVersion).Version
	switch {
	case strings.HasPrefix(ciliumVersion, "1.6"):
		return cilium_manifests.PreflightManifestv16
	case strings.HasPrefix(ciliumVersion, "1.7"):
		return ""
	}
	return ""
}

type ciliumCallbacks struct{}

func (ciliumCallbacks) beforeApply(client clientset.Interface, addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	ciliumVersion := kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, addonConfiguration.ClusterVersion).Version
	_, config, err := kubernetes.GetAdminClientSetWithConfig()
	if err != nil {
		return errors.Wrap(err, "unable to get admin client set")
	}
	if err := cni.CreateCiliumSecret(client, ciliumVersion); err != nil {
		return err
	}
	if err := cni.CreateOrUpdateCiliumConfigMap(client, ciliumVersion); err != nil {
		return err
	}

	// Handle migration from etcd to CRD when upgrading from Cilium 1.5.
	needsMigration, err := cni.NeedsEtcdToCrdMigration(client, ciliumVersion)
	if err != nil {
		return err
	}
	if !needsMigration {
		return nil
	}
	if err := cni.MigrateEtcdToCrd(client, config); err != nil {
		return err
	}
	if err := cni.RemoveEtcdConfig(client); err != nil {
		return err
	}

	return nil
}

func (ciliumCallbacks) afterApply(client clientset.Interface, addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	return nil
}
