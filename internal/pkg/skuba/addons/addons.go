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

package addons

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
)

type addonPriority uint

const (
	highPriority   addonPriority = iota
	normalPriority addonPriority = iota
)

var Addons = map[kubernetes.Addon]Addon{}

type Addon struct {
	addon         kubernetes.Addon
	templater     addonTemplater
	callbacks     addonCallbacks
	addonPriority addonPriority
}

type addonCallbacks interface {
	beforeApply(AddonConfiguration, *skuba.SkubaConfiguration) error
	afterApply(AddonConfiguration, *skuba.SkubaConfiguration) error
}

type AddonConfiguration struct {
	ClusterVersion *version.Version
	ControlPlane   string
	ClusterName    string
}

type addonTemplater func(AddonConfiguration) string

type ApplyBehavior uint

const (
	// This is the default behavior for all operations except for Bootstrap,
	// the addon is always re-rendered prior to being applied. In an addons
	// upgrade operation for example, we always want to re-render the latest
	// contents and never reuse local file contents in case the upgrade was
	// executed inside a cluster definition folder
	AlwaysRender ApplyBehavior = iota
	// This is the desired behavior for Bootstrap, when the user can tweak
	// the addon configurations after `skuba cluster init`, so
	// `skuba node bootstrap` will apply the modified settings instead of
	// re-rendering them forcefully
	SkipRenderIfConfigFilePresent ApplyBehavior = iota
)

type renderContext struct {
	addon  Addon
	config AddonConfiguration
}

func (renderContext renderContext) AnnotatedVersion() string {
	return fmt.Sprintf("addon.caasp.suse.com/manifest-version: \"%s\"", renderContext.ManifestVersion())
}

func (renderContext renderContext) ManifestVersion() string {
	addonVersion := kubernetes.AddonVersionForClusterVersion(renderContext.addon.addon, renderContext.config.ClusterVersion)
	if addonVersion == nil {
		return ""
	}
	return fmt.Sprintf("%s-%d", addonVersion.Version, addonVersion.ManifestVersion)
}

func registerAddon(addon kubernetes.Addon, addonTemplater addonTemplater, callbacks addonCallbacks, addonPriority addonPriority) {
	Addons[addon] = Addon{
		addon:         addon,
		templater:     addonTemplater,
		callbacks:     callbacks,
		addonPriority: addonPriority,
	}
}

func addonsByPriority() []Addon {
	sortedAddons := make([]Addon, len(Addons))
	i := 0
	for _, addon := range Addons {
		sortedAddons[i] = addon
		i++
	}
	sort.Slice(sortedAddons, func(i, j int) bool {
		return sortedAddons[i].addonPriority < sortedAddons[j].addonPriority
	})
	return sortedAddons
}

func DeployAddons(client clientset.Interface, addonConfiguration AddonConfiguration, applyBehavior ApplyBehavior) error {
	skubaConfiguration, err := skuba.GetSkubaConfiguration(client)
	if err != nil {
		return err
	}
	for _, addon := range addonsByPriority() {
		addonName := addon.addon
		if !addon.IsPresentForClusterVersion(addonConfiguration.ClusterVersion) {
			// This registered addon is not available on the chosen Kubernetes version, skip it
			continue
		}
		hasToBeApplied, err := addon.HasToBeApplied(addonConfiguration, skubaConfiguration)
		if err != nil {
			klog.Errorf("cannot determine if %q addon needs to be applied, skipping...", addonName)
			continue
		}
		if hasToBeApplied {
			if err := addon.Apply(addonConfiguration, skubaConfiguration, applyBehavior); err == nil {
				klog.V(1).Infof("%q addon correctly applied", addonName)
			} else {
				klog.Errorf("failed to apply %q addon (%v)", addonName, err)
			}
		} else {
			klog.V(1).Infof("skipping %q addon apply", addonName)
		}
	}
	return nil
}

func (addon Addon) Render(addonConfiguration AddonConfiguration) (string, error) {
	template, err := template.New("").Parse(addon.templater(addonConfiguration))
	if err != nil {
		return "", errors.Wrap(err, "could not parse manifest template")
	}
	var rendered bytes.Buffer
	err = template.Execute(&rendered, renderContext{
		addon:  addon,
		config: addonConfiguration,
	})
	if err != nil {
		return "", errors.Wrap(err, "could not render manifest template")
	}
	return rendered.String(), nil
}

func (addon Addon) IsPresentForClusterVersion(clusterVersion *version.Version) bool {
	return kubernetes.AddonVersionForClusterVersion(addon.addon, clusterVersion) != nil
}

func (addon Addon) HasToBeApplied(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) (bool, error) {
	if !addon.IsPresentForClusterVersion(addonConfiguration.ClusterVersion) {
		// TODO (ereslibre): this logic can be triggered if some registered
		//       addons are not supported in all Kubernetes versions. Either:
		//
		//   a) When rendering all addons on `skuba cluster init`, we skip those
		//      that don't apply to the chosen Kubernetes version.
		//
		//   b) When running `skuba addon upgrade apply`; in this case (hence the
		//      TODO), should we trigger a deletion of the addons that are not present
		//      in the new version but were present on the old Kubernetes version? For
		//      now, just return that it doesn't have to be applied.
		return false, nil
	}
	if skubaConfiguration.AddonsVersion == nil {
		return true, nil
	}
	currentAddonVersion, found := skubaConfiguration.AddonsVersion[addon.addon]
	if !found {
		return true, nil
	}
	addonVersion := kubernetes.AddonVersionForClusterVersion(addon.addon, addonConfiguration.ClusterVersion)
	return addonVersionLower(currentAddonVersion, addonVersion), nil
}

func (addon Addon) needsRender(applyBehavior ApplyBehavior) bool {
	if applyBehavior == AlwaysRender {
		return true
	}
	_, err := os.Stat(addon.manifestPath())
	return err != nil
}

func (addon Addon) addonPath() string {
	return filepath.Join(skubaconstants.AddonsDir(), string(addon.addon))
}

func (addon Addon) manifestPath() string {
	return filepath.Join(addon.addonPath(), fmt.Sprintf("%s.yaml", addon.addon))
}

func (addon Addon) Write(addonConfiguration AddonConfiguration) error {
	addonManifest, err := addon.Render(addonConfiguration)
	if err != nil {
		return errors.Wrapf(err, "unable to render %s addon template", addon.addon)
	}
	if err := os.MkdirAll(addon.addonPath(), 0700); err != nil {
		return errors.Wrapf(err, "unable to create folder for addon %s", addon.addon)
	}
	if err := ioutil.WriteFile(addon.manifestPath(), []byte(addonManifest), 0600); err != nil {
		return errors.Wrapf(err, "unable to write %s addon rendered template", addon.addon)
	}
	return nil
}

func (addon Addon) Apply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration, applyBehavior ApplyBehavior) error {
	klog.V(1).Infof("applying %q addon", addon.addon)
	if addon.callbacks != nil {
		if err := addon.callbacks.beforeApply(addonConfiguration, skubaConfiguration); err != nil {
			klog.Errorf("failed on %q addon BeforeApply callback: %v", addon.addon, err)
			return err
		}
	}
	var renderedManifest string
	if addon.needsRender(applyBehavior) {
		var err error
		renderedManifest, err = addon.Render(addonConfiguration)
		if err != nil {
			return err
		}
	} else {
		renderedManifestBytes, err := ioutil.ReadFile(addon.manifestPath())
		if err != nil {
			return err
		}
		renderedManifest = string(renderedManifestBytes)
	}
	cmd := exec.Command("kubectl", "apply", "--kubeconfig", skubaconstants.KubeConfigAdminFile(), "-f", "-")
	cmd.Stdin = bytes.NewBuffer([]byte(renderedManifest))
	if combinedOutput, err := cmd.CombinedOutput(); err != nil {
		klog.Errorf("failed to run kubectl apply: %s", combinedOutput)
		return err
	}
	if addon.callbacks != nil {
		if err := addon.callbacks.afterApply(addonConfiguration, skubaConfiguration); err != nil {
			// TODO: should we rollback here?
			klog.Errorf("failed on %q addon AfterApply callback: %v", addon.addon, err)
			return err
		}
	}
	return updateSkubaConfigMapWithAddonVersion(addon.addon, addonConfiguration.ClusterVersion, skubaConfiguration)
}

func addonVersionLower(current *kubernetes.AddonVersion, updated *kubernetes.AddonVersion) bool {
	// If we don't have a version to compare to, assume it's not lower
	if current == nil {
		return false
	}
	return current.ManifestVersion < updated.ManifestVersion
}

func updateSkubaConfigMapWithAddonVersion(addon kubernetes.Addon, clusterVersion *version.Version, skubaConfiguration *skuba.SkubaConfiguration) error {
	addonVersion := kubernetes.AddonVersionForClusterVersion(addon, clusterVersion)
	if skubaConfiguration.AddonsVersion == nil {
		skubaConfiguration.AddonsVersion = map[kubernetes.Addon]*kubernetes.AddonVersion{}
	}
	skubaConfiguration.AddonsVersion[addon] = addonVersion
	return skuba.UpdateSkubaConfiguration(skubaConfiguration)
}
