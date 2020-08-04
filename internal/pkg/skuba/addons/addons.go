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

// Package addons provides the mechanism to extend the kubernetes functionality by applying
// addons that provide new functions. This package also includes the addons
package addons

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

const (
	addonTemplateWarning = `# Do not edit this file directly.
#
# Any manual changes made to this file will be ignored.
#
# If you want to adapt this addon manifest, use the "patches" directory,
# that allows you to provide strategic merge patches, and json 6902 patches
# (https://tools.ietf.org/html/rfc6902).
#
# More information: https://github.com/kubernetes-sigs/kustomize/blob/master/docs/glossary.md
`

	kustomizeTemplate = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
{{ range .Resources -}}
  - base/{{ . }}
{{ end -}}
{{ if gt (len .Patches) 0 -}}
patches:
{{ range .Patches -}}
  - patches/{{ . }}
{{ end -}}
{{ end -}}
`
)

const (
	CniAddOn     AddOnType = "CNI"
	GenericAddOn AddOnType = "GENERIC"
)

var Addons = map[kubernetes.Addon]Addon{}

type AddOnType string

type Addon struct {
	Addon              kubernetes.Addon
	templater          addonTemplater
	preflightTemplater preflightAddonTemplater
	callbacks          addonCallbacks
	addonPriority      addonPriority
	getImageCallbacks  []getImageCallback
	AddOnType          AddOnType
}

type addonCallbacks interface {
	beforeApply(clientset.Interface, AddonConfiguration, *skuba.SkubaConfiguration) error
	afterApply(clientset.Interface, AddonConfiguration, *skuba.SkubaConfiguration) error
}

type AddonConfiguration struct {
	ClusterVersion *version.Version
	ControlPlane   string
	ClusterName    string
}

type addonTemplater func(AddonConfiguration) string
type preflightAddonTemplater func(AddonConfiguration) string
type getImageCallback func(clusterVersion *version.Version, imageTag string) string

type renderContext struct {
	addon  Addon
	config AddonConfiguration
}

func (renderContext renderContext) AnnotatedVersion() string {
	return fmt.Sprintf("addon.caasp.suse.com/manifest-version: \"%s\"", renderContext.ManifestVersion())
}

func (renderContext renderContext) ManifestVersion() string {
	addonVersion := kubernetes.AddonVersionForClusterVersion(renderContext.addon.Addon, renderContext.config.ClusterVersion)
	if addonVersion == nil {
		return ""
	}
	return fmt.Sprintf("%s-%d", addonVersion.Version, addonVersion.ManifestVersion)
}

// registerAddon incorporates one addon information to the Addons map that keeps track of the
// addons which will get deployed
func registerAddon(addon kubernetes.Addon, addonType AddOnType, addonTemplater addonTemplater, preflightAddonTemplater preflightAddonTemplater, callbacks addonCallbacks, addonPriority addonPriority, getImageCallbacks []getImageCallback) {
	Addons[addon] = Addon{
		Addon:              addon,
		templater:          addonTemplater,
		preflightTemplater: preflightAddonTemplater,
		callbacks:          callbacks,
		addonPriority:      addonPriority,
		getImageCallbacks:  getImageCallbacks,
		AddOnType:          addonType,
	}
}

// addonsByPriority sorts the addons in the Addons map by their priority set by the
// addon.addonPriority uint and returns a slice
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

func CheckLocalAddonsBaseManifests(addonConfiguration AddonConfiguration) (bool, error) {
	for _, addon := range addonsByPriority() {
		if !addon.IsPresentForClusterVersion(addonConfiguration.ClusterVersion) {
			// This registered addon is not available on the chosen Kubernetes version, skip it
			continue
		}
		if match, err := addon.compareLocalBaseManifest(addonConfiguration); err != nil || !match {
			return match, err
		}
	}
	return true, nil
}

// DeployAddons loops over the sorted list of addons, checks if each needs to be deployed and
// triggers its deployment
func DeployAddons(client clientset.Interface, addonConfiguration AddonConfiguration, dryRun bool) error {
	skubaConfiguration, err := skuba.GetSkubaConfiguration(client)
	if err != nil {
		return err
	}
	for _, addon := range addonsByPriority() {
		addonName := addon.Addon
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
			if err := addon.Apply(client, addonConfiguration, skubaConfiguration, dryRun); err == nil {
				klog.V(1).Infof("%q addon correctly applied", addonName)
			} else {
				klog.Errorf("failed to apply %q addon (%v)", addonName, err)
				return err
			}
		} else {
			klog.V(1).Infof("skipping %q addon apply", addonName)
		}
	}
	return nil
}

func (addon Addon) renderTemplate(template *template.Template, addonConfiguration AddonConfiguration) (string, error) {
	var rendered bytes.Buffer
	if err := template.Execute(&rendered, renderContext{
		addon:  addon,
		config: addonConfiguration,
	}); err != nil {
		return "", errors.Wrap(err, "could not render manifest template")
	}
	return rendered.String(), nil
}

// Render substitutes the variables in the template and returns a string with the addon
// manifest ready
func (addon Addon) Render(addonConfiguration AddonConfiguration) (string, error) {
	template, err := template.New("").Parse(addon.templater(addonConfiguration))
	if err != nil {
		return "", errors.Wrap(err, "could not parse manifest template")
	}
	return addon.renderTemplate(template, addonConfiguration)
}

func (addon Addon) RenderPreflight(addonConfiguration AddonConfiguration) (string, error) {
	template, err := template.New("").Parse(addon.preflightTemplater(addonConfiguration))
	if err != nil {
		return "", errors.Wrap(err, "could not parse preflight template")
	}
	return addon.renderTemplate(template, addonConfiguration)
}

// IsPresentForClusterVersion verifies if the Addon can be deployed with the current k8s version
func (addon Addon) IsPresentForClusterVersion(clusterVersion *version.Version) bool {
	return kubernetes.AddonVersionForClusterVersion(addon.Addon, clusterVersion) != nil
}

// HasToBeApplied decides if the Addon is deployed by checking its version with addonVersionLower
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
	// Check whether this is a CNI addon and whether its base config has been rendered.
	// If it's a CNI plugin and the base config has not been rendered, we can assume
	// the user requested a different CNI plugin and this addon does not need to be
	// applied.
	if info, err := os.Stat(addon.addonDir()); addon.AddOnType == CniAddOn && (os.IsNotExist(err) || !info.IsDir()) {
		return false, nil
	}
	if skubaConfiguration.AddonsVersion == nil {
		return true, nil
	}
	currentAddonVersion, found := skubaConfiguration.AddonsVersion[addon.Addon]
	if !found {
		return true, nil
	}
	addonVersion := kubernetes.AddonVersionForClusterVersion(addon.Addon, addonConfiguration.ClusterVersion)
	return addonVersionLower(currentAddonVersion, addonVersion), nil
}

func (addon Addon) addonDir() string {
	return filepath.Join(skubaconstants.AddonsDir(), string(addon.Addon))
}

func (addon Addon) baseResourcesDir(rootDir string) string {
	return filepath.Join(rootDir, "base")
}

func (addon Addon) patchResourcesDir(rootDir string) string {
	return filepath.Join(rootDir, "patches")
}

func (addon Addon) manifestFilename() string {
	return fmt.Sprintf("%s.yaml", addon.Addon)
}

func (addon Addon) preflightManifestFilename() string {
	return fmt.Sprintf("%s-preflight.yaml", addon.Addon)
}

func (addon Addon) legacyManifestPath(rootDir string) string {
	return filepath.Join(rootDir, addon.manifestFilename())
}

func (addon Addon) manifestPath(rootDir string) string {
	return filepath.Join(addon.baseResourcesDir(rootDir), addon.manifestFilename())
}

func (addon Addon) preflightManifestPath(rootDir string) string {
	return filepath.Join(addon.baseResourcesDir(rootDir), addon.preflightManifestFilename())
}

func (addon Addon) kustomizeContents(resourceManifests []string, patchManifests []string) (string, error) {
	template, err := template.New("").Parse(kustomizeTemplate)
	if err != nil {
		return "", err
	}
	var rendered bytes.Buffer
	err = template.Execute(&rendered, struct {
		Resources []string
		Patches   []string
	}{
		Resources: resourceManifests,
		Patches:   patchManifests,
	})
	if err != nil {
		return "", errors.Wrap(err, "could not render configuration")
	}
	return rendered.String(), nil
}

func (addon Addon) kustomizePath(rootDir string) string {
	return filepath.Join(rootDir, "kustomization.yaml")
}

func (addon Addon) compareLocalBaseManifest(addonConfiguration AddonConfiguration) (bool, error) {
	localManifest, err := ioutil.ReadFile(addon.manifestPath(addon.addonDir()))
	if err != nil {
		return false, errors.Wrapf(err, "unable to read %s addon rendered template", addon.Addon)
	}

	addonManifest, err := addon.Render(addonConfiguration)
	if err != nil {
		return false, errors.Wrapf(err, "unable to render %s addon template", addon.Addon)
	}

	if !reflect.DeepEqual(localManifest, []byte(addonTemplateWarning+addonManifest)) {
		return false, nil
	}
	return true, nil
}

// Write creates the manifest yaml file of the Addon after rendering its template
func (addon Addon) Write(addonConfiguration AddonConfiguration) error {
	addonManifest, err := addon.Render(addonConfiguration)
	if err != nil {
		return errors.Wrapf(err, "unable to render %s addon template", addon.Addon)
	}
	baseResourcesDir := addon.baseResourcesDir(addon.addonDir())
	if err := os.MkdirAll(baseResourcesDir, 0700); err != nil {
		return errors.Wrapf(err, "unable to create directory: %s", baseResourcesDir)
	}
	patchResourcesDir := addon.patchResourcesDir(addon.addonDir())
	if err := os.MkdirAll(patchResourcesDir, 0700); err != nil {
		return errors.Wrapf(err, "unable to create directory: %s", patchResourcesDir)
	}

	// migrates legacy addon manifest if existed
	legacyManifestPath := addon.legacyManifestPath(addon.addonDir())
	if f, err := os.Stat(legacyManifestPath); !os.IsNotExist(err) && !f.IsDir() {
		if err := os.Remove(legacyManifestPath); err != nil {
			return errors.Wrapf(err, "unable to remove %s addon legacy rendered template", addon.Addon)
		}
	}

	if err := ioutil.WriteFile(addon.manifestPath(addon.addonDir()), []byte(addonTemplateWarning+addonManifest), 0600); err != nil {
		return errors.Wrapf(err, "unable to write %s addon rendered template", addon.Addon)
	}
	return nil
}

// applyPreflight applies the preflight deployment/daemonset manifest if such
// is defined by the addon. It returns a bool whether preflight was defined
// by the addon and an error.
func (addon Addon) applyPreflight(addonConfiguration AddonConfiguration, rootDir string, dryRun bool) (bool, error) {
	renderedPreflightManifest, err := addon.RenderPreflight(addonConfiguration)
	if err != nil {
		return false, err
	}
	if renderedPreflightManifest == "" {
		return false, nil
	}
	preflightManifestPath := addon.preflightManifestPath(rootDir)
	if err := ioutil.WriteFile(preflightManifestPath, []byte(renderedPreflightManifest), 0600); err != nil {
		return true, errors.Wrapf(err, "could not create %q addon manifests", addon.Addon)
	}
	kubectlArgs := []string{
		"apply",
		"--kubeconfig", skubaconstants.KubeConfigAdminFile(),
		"-f", preflightManifestPath,
	}
	if dryRun {
		kubectlArgs = append(kubectlArgs, "--dry-run=server")
	}
	cmd := exec.Command("kubectl", kubectlArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// prints out stderr
		fmt.Printf("%s", stderr.Bytes())
		return true, err
	}
	return true, nil
}

func (addon Addon) deletePreflight(rootDir string) error {
	preflightManifestPath := addon.preflightManifestPath(rootDir)
	cmd := exec.Command("kubectl", "delete", "--kubeconfig", skubaconstants.KubeConfigAdminFile(), "-f", preflightManifestPath)
	if combinedOutput, err := cmd.CombinedOutput(); err != nil {
		klog.Errorf("failed to run kubectl delete: %s", combinedOutput)
		return err
	}
	return nil
}

// Apply deploys the addon by calling kubectl apply and pointing to the generated addon
// manifest
func (addon Addon) Apply(client clientset.Interface, addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration, dryRun bool) error {
	klog.V(1).Infof("applying %q addon", addon.Addon)

	sandboxDir, err := addon.createSandbox()
	if err != nil {
		return errors.Wrapf(err, "could not create %q addon sandbox", addon.Addon)
	}
	defer func() {
		_ = addon.deleteSandbox(sandboxDir)
	}()
	baseResourcesDir := addon.baseResourcesDir(sandboxDir)
	if err := os.MkdirAll(baseResourcesDir, 0700); err != nil {
		return errors.Wrapf(err, "unable to create directory: %s", baseResourcesDir)
	}

	hasPreflight := false
	if addon.preflightTemplater != nil {
		hasPreflight, err = addon.applyPreflight(addonConfiguration, addon.addonDir(), dryRun)
		if err != nil {
			return err
		}
	}
	if addon.callbacks != nil && !dryRun {
		if err = addon.callbacks.beforeApply(client, addonConfiguration, skubaConfiguration); err != nil {
			klog.Errorf("failed on %q addon BeforeApply callback: %v", addon.Addon, err)
			return err
		}
	}
	if hasPreflight && !dryRun {
		// since we did not run applyPreflight if dryRun=true
		// bypass delete preflight manifests
		if err = addon.deletePreflight(sandboxDir); err != nil {
			return err
		}
	}

	renderedManifest, err := addon.Render(addonConfiguration)
	if err != nil {
		return err
	}

	patchResourcesDir := addon.patchResourcesDir(sandboxDir)
	if err := os.MkdirAll(patchResourcesDir, 0700); err != nil {
		return errors.Wrapf(err, "unable to create directory: %s", patchResourcesDir)
	}
	// Write base resources
	if err := ioutil.WriteFile(addon.manifestPath(sandboxDir), []byte(renderedManifest), 0600); err != nil {
		return errors.Wrapf(err, "could not create %q addon manifests", addon.Addon)
	}
	// Write patch resources
	if err := addon.copyCustomizePatches(sandboxDir); err != nil {
		return errors.Wrapf(err, "could not link %q addon patches", addon.Addon)
	}

	patchList, err := addon.listPatches()
	if err != nil {
		return errors.Wrapf(err, "could not list patches for %q addon", addon.Addon)
	}
	kustomizeContents, err := addon.kustomizeContents([]string{addon.manifestFilename()}, patchList)
	if err != nil {
		return errors.Wrapf(err, "could not render kustomize file")
	}
	if err = ioutil.WriteFile(addon.kustomizePath(sandboxDir), []byte(kustomizeContents), 0600); err != nil {
		return errors.Wrapf(err, "could not create %q kustomize file", addon.Addon)
	}
	kubectlArgs := []string{
		"apply",
		"--kubeconfig", skubaconstants.KubeConfigAdminFile(),
		"-k", sandboxDir,
	}
	if dryRun {
		kubectlArgs = append(kubectlArgs, "--dry-run=server")
	}
	cmd := exec.Command("kubectl", kubectlArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// prints out stderr
		fmt.Printf("%s", stderr.Bytes())
		return err
	}
	if addon.callbacks != nil && !dryRun {
		if err = addon.callbacks.afterApply(client, addonConfiguration, skubaConfiguration); err != nil {
			// TODO: should we rollback here?
			klog.Errorf("failed on %q addon AfterApply callback: %v", addon.Addon, err)
			return err
		}
	}

	if dryRun {
		// immediately return, do not update skuba-config ConfigMap
		return nil
	}
	return updateSkubaConfigMapWithAddonVersion(client, addon.Addon, addonConfiguration.ClusterVersion, skubaConfiguration)
}

// Images returns the images required for this Addon to properly function
func (addon Addon) Images(clusterVersion *version.Version, imageTag string) []string {
	images := []string{}
	for _, cb := range addon.getImageCallbacks {
		images = append(images, cb(clusterVersion, imageTag))
	}
	return images
}

func (addon Addon) listPatches() ([]string, error) {
	result := []string{}
	sourcePatches, err := filepath.Glob(filepath.Join(addon.patchResourcesDir(addon.addonDir()), "*.yaml"))
	if err != nil {
		return result, err
	}
	for _, patchPath := range sourcePatches {
		result = append(result, filepath.Base(patchPath))
	}
	return result, nil
}

func (addon Addon) createSandbox() (string, error) {
	return ioutil.TempDir("", fmt.Sprintf("skuba-addon-%s", addon.Addon))
}

func (addon Addon) deleteSandbox(sandboxDir string) error {
	return os.RemoveAll(sandboxDir)
}

func (addon Addon) copyCustomizePatches(targetDir string) error {
	patches, err := addon.listPatches()
	if err != nil {
		return err
	}
	for _, patch := range patches {
		source, err := os.Open(filepath.Join(addon.patchResourcesDir(addon.addonDir()), patch))
		if err != nil {
			return err
		}
		target, err := os.Create(filepath.Join(addon.patchResourcesDir(targetDir), patch))
		if err != nil {
			return err
		}
		if _, err := io.Copy(target, source); err != nil {
			return err
		}
	}
	return nil
}

// addonVersionLower checks if the updated version of the Addon is greater than the current
func addonVersionLower(current *kubernetes.AddonVersion, updated *kubernetes.AddonVersion) bool {
	// If we don't have a version to compare to, assume it's not lower
	if current == nil {
		return false
	}
	return current.ManifestVersion < updated.ManifestVersion
}

// updateSkubaConfigMapWithAddonVersion updates the general Skuba config to include the
// information of the Addon which was deployed
func updateSkubaConfigMapWithAddonVersion(client clientset.Interface, addon kubernetes.Addon, clusterVersion *version.Version, skubaConfiguration *skuba.SkubaConfiguration) error {
	addonVersion := kubernetes.AddonVersionForClusterVersion(addon, clusterVersion)
	if skubaConfiguration.AddonsVersion == nil {
		skubaConfiguration.AddonsVersion = map[kubernetes.Addon]*kubernetes.AddonVersion{}
	}
	skubaConfiguration.AddonsVersion[addon] = addonVersion
	return skuba.UpdateSkubaConfiguration(client, skubaConfiguration)
}
