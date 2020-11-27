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

package ssh

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
)

type KubernetesUploadSecretsErrorBehavior uint

const (
	KubernetesUploadSecretsFailOnError     KubernetesUploadSecretsErrorBehavior = iota
	KubernetesUploadSecretsContinueOnError KubernetesUploadSecretsErrorBehavior = iota
)

func init() {
	stateMap["kubernetes.bootstrap.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsContinueOnError)
	stateMap["kubernetes.join.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsFailOnError)
	stateMap["kubernetes.install-fresh-pkgs"] = kubernetesFreshInstallAllPkgs
	stateMap["kubernetes.upgrade-stage-one"] = kubernetesUpgradeStageOne
	stateMap["kubernetes.upgrade-stage-two"] = kubernetesUpgradeStageTwo
	stateMap["kubernetes.restart-services"] = kubernetesRestartServices
	stateMap["kubernetes.enable-services"] = kubernetesEnsureServicesEnabled
}

func kubernetesUploadSecrets(errorHandling KubernetesUploadSecretsErrorBehavior) Runner {
	return func(t *Target, data interface{}) error {
		for _, file := range deployments.Secrets {
			f, err := os.Stat(file)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				if errorHandling == KubernetesUploadSecretsFailOnError {
					return err
				}
			}
			if err := t.target.UploadFile(file, filepath.Join(constants.KubernetesDir, file), f.Mode()); err != nil {
				if errorHandling == KubernetesUploadSecretsFailOnError {
					return err
				}
			}
		}
		return nil
	}
}

// kubernetesParseInterfaceVersions parses the versions given in a consistent way for the fresh
// install and the different upgrade stages, returning first the current version, then optionally the updated version.
func kubernetesParseInterfaceVersions(data interface{}) (string, string, error) {
	kubernetesBaseOSConfiguration, ok := data.(deployments.KubernetesBaseOSConfiguration)
	if !ok {
		return "", "", errors.New("couldn't access kubernetes base OS configuration")
	}

	v, err := version.ParseSemantic(kubernetesBaseOSConfiguration.CurrentVersion)
	if err != nil {
		return "", "", err
	}
	currentVersion := kubernetes.MajorMinorVersion(v)
	updatedVersion := ""

	if kubernetesBaseOSConfiguration.UpdatedVersion != "" {
		updatedVersion = kubernetes.MajorMinorVersion(version.MustParseSemantic(kubernetesBaseOSConfiguration.UpdatedVersion))
	}
	return currentVersion, updatedVersion, nil
}

func kubernetesFreshInstallAllPkgs(t *Target, data interface{}) error {
	// This only applies on new nodes (bootstrap/join)
	current, _, err := kubernetesParseInterfaceVersions(data)
	if err != nil {
		return err
	}
	var pkgs []string

	// Standard packages for a new cluster
	pkgs = append(pkgs, "+caasp-release", "skuba-update", "supportutils-plugin-suse-caasp")
	// Version specific
	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-kubeadm", current))
	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-kubelet", current))
	pkgs = append(pkgs, fmt.Sprintf("+kubernetes%s-client", current))
	pkgs = append(pkgs, fmt.Sprintf("+cri-o-%s*", current))
	pkgs = append(pkgs, fmt.Sprintf("+cri-tools-%s*", current))

	_, _, err = t.ZypperInstall(pkgs...)
	return err
}

func kubernetesUpgradeStageOne(t *Target, data interface{}) error {
	// Stage1 upgrades only kubeadm.
	// zypper install -- -kubernetes-old-kubeadm +kubernetes-new-kubeadm
	currentV, nextV, err := kubernetesParseInterfaceVersions(data)
	if err != nil {
		return err
	}
	if nextV == "" {
		return errors.New("Incorrect upgrade version")
	}

	var pkgs []string

	if currentV == skubaconstants.LastCaaSP4KubernetesVersion {
		// 1.17 is the last version included in CaaSP4. It's the tipping
		// point where we changed our packaging.
		// On 1.17 we can't remove kubernetes-1.17-kubeadm, because it doesn't exist.
		// We are removing kubeadm while keeping kubelet alive to its version 1.17.
		// For the initial migration we need to update crio kubeadm
		// to 1.18 in stage1, due to conflict resolution: the caasp4
		// cri-o-kubeadm-criconfig requires kubernetes-kubeadm which is
		// not provided anymore (when we remove kubernetes-kubeadm, and
		// because we don't want to have the same provides: on the new
		// package to avoid upgrade during zypper migration).
		// we need to remove cri-o in stage2 else 1.17 kubelet could
		// complain about cri-runtime being absent.
		pkgs = append(pkgs, fmt.Sprintf("-patterns-caasp-Node-%s", skubaconstants.LastCaaSP4KubernetesVersion))
		pkgs = append(pkgs, fmt.Sprintf("-\"kubernetes-kubeadm<%s\"", skubaconstants.FirstCaaSP5KubernetesVersion))
		pkgs = append(pkgs, "-caasp-config", "-cri-o-kubeadm-criconfig")
		// We also need to install cri-o-1.18-kubeadm-criconfig at the same time
		// that we remove cri-o-kubeadm-criconfig, because otherwise the kubernetes-1.18-kubeadm requirements would trigger the installation of any cri-o-*-kubeadm-criconfig
		// leading to unexpected results on late migrations to SP2 and therefore late upgrades to caasp 4.5
		pkgs = append(pkgs, fmt.Sprintf("cri-o-%s-kubeadm-criconfig", skubaconstants.FirstCaaSP5KubernetesVersion))
	} else {
		pkgs = append(pkgs, fmt.Sprintf("-kubernetes-%s-kubeadm", currentV))
	}

	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-kubeadm", nextV))
	_, _, err = t.ZypperInstall(pkgs...)
	return err
}

func kubernetesUpgradeStageTwo(t *Target, data interface{}) error {
	// Stage2 installs the rest of the packages during the upgrade
	// with zypper install -- -<previous>-<component> +<next>-<component>
	currentV, nextV, err := kubernetesParseInterfaceVersions(data)
	if err != nil {
		return err
	}
	if nextV == "" {
		return errors.New("Incorrect upgrade version")
	}

	var pkgs []string

	if currentV == skubaconstants.LastCaaSP4KubernetesVersion {
		// on 1.17 we need to finalize the cleanup for the
		// caasp4 to 4.5 migration during this stage.
		pkgs = append(pkgs, fmt.Sprintf("-\"kubernetes-kubelet<%s\"", skubaconstants.FirstCaaSP5KubernetesVersion))
		pkgs = append(pkgs, "-kubernetes-common")
		pkgs = append(pkgs, fmt.Sprintf("-\"kubernetes-client<%s\"", skubaconstants.FirstCaaSP5KubernetesVersion))
		pkgs = append(pkgs, fmt.Sprintf("-\"cri-o<%s\"", skubaconstants.FirstCaaSP5KubernetesVersion))
		pkgs = append(pkgs, fmt.Sprintf("-\"cri-tools<%s\"", skubaconstants.FirstCaaSP5KubernetesVersion))
	} else {
		pkgs = append(pkgs, fmt.Sprintf("-kubernetes%s-*", currentV))
		pkgs = append(pkgs, fmt.Sprintf("-cri-o-%s*", currentV))
		pkgs = append(pkgs, fmt.Sprintf("-cri-tools-%s*", currentV))
	}

	pkgs = append(pkgs, fmt.Sprintf("+kubernetes%s-client", nextV))
	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-kubelet", nextV))
	pkgs = append(pkgs, fmt.Sprintf("+cri-o-%s*", nextV))
	pkgs = append(pkgs, fmt.Sprintf("+cri-tools-%s*", nextV))
	_, _, err = t.ZypperInstall(pkgs...)
	return err
}

func kubernetesRestartServices(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "restart", "crio", "kubelet")
	return err
}

func kubernetesEnsureServicesEnabled(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "enable", "crio", "kubelet")
	return err
}
