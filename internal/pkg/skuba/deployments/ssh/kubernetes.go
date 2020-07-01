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
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
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
}

func kubernetesUploadSecrets(errorHandling KubernetesUploadSecretsErrorBehavior) Runner {
	return func(t *Target, data interface{}) error {
		for _, file := range deployments.Secrets {
			if err := t.target.UploadFile(file, filepath.Join(constants.KubernetesDir, file)); err != nil {
				if errorHandling == KubernetesUploadSecretsFailOnError {
					return err
				}
			}
		}
		return nil
	}
}

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
	current, _, err := kubernetesParseInterfaceVersions(data)
	if err != nil {
		return err
	}
	var pkgs []string

	// Standard packages for a new cluster
	pkgs = append(pkgs, "+caasp-release", "skuba-update", "supportutils-plugin-suse-caasp", "cri-tools")
	// Version specific
	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-kubeadm", current))
	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-kubelet", current))
	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-client", current))
	pkgs = append(pkgs, fmt.Sprintf("+cri-o-%s*", current))

	_, _, err = t.zypperInstall(pkgs...)
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

	if currentV == "1.17" {
		// 1.17 is the last version included in CaaSP4. It's the tipping
		// point where we changed our packaging.
		// On 1.17 we can't remove kubernetes-1.17-kubeadm, because it doesn't exist.
		// Removing kubeadm keeps kubelet alive.
		// The rest needs to be removed on the next stage.
		pkgs = append(pkgs, "-patterns-caasp-Node-1.17", "-\"kubernetes-kubeadm<1.18\"", "-caasp-config")
	} else {
		pkgs = append(pkgs, fmt.Sprintf("-kubernetes-%s-kubeadm", currentV))
	}

	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-kubeadm", nextV))
	_, _, err = t.zypperInstall(pkgs...)
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

	if currentV == "1.17" {
		// on 1.17 we need to finalize the cleanup for the
		// caasp4 to 5 migration
		pkgs = append(pkgs, "-kubernetes-kubelet")
		pkgs = append(pkgs, "-kubernetes-common")
		pkgs = append(pkgs, "-kubernetes-client")
		pkgs = append(pkgs, "-cri-o*")
	} else {
		pkgs = append(pkgs, fmt.Sprintf("-kubernetes-%s-*", currentV))
		pkgs = append(pkgs, fmt.Sprintf("-cri-o-%s*", currentV))
	}

	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-client", nextV))
	pkgs = append(pkgs, fmt.Sprintf("+kubernetes-%s-kubelet", nextV))
	pkgs = append(pkgs, fmt.Sprintf("+cri-o-%s*", nextV))
	_, _, err = t.zypperInstall(pkgs...)
	return err
}

func kubernetesRestartServices(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "restart", "crio", "kubelet")
	return err
}
