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

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/pkg/errors"
)

type KubernetesUploadSecretsErrorBehavior uint

const (
	KubernetesUploadSecretsFailOnError     KubernetesUploadSecretsErrorBehavior = iota
	KubernetesUploadSecretsContinueOnError KubernetesUploadSecretsErrorBehavior = iota
)

func init() {
	stateMap["kubernetes.bootstrap.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsContinueOnError)
	stateMap["kubernetes.join.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsFailOnError)
	stateMap["kubernetes.install-base-packages"] = kubernetesInstallBasePackages
	stateMap["kubernetes.restart-services"] = kubernetesRestartServices
}

func kubernetesUploadSecrets(errorHandling KubernetesUploadSecretsErrorBehavior) Runner {
	return func(t *Target, data interface{}) error {
		for _, file := range deployments.Secrets {
			if err := t.target.UploadFile(file, filepath.Join("/etc/kubernetes", file)); err != nil {
				if errorHandling == KubernetesUploadSecretsFailOnError {
					return err
				}
			}
		}
		return nil
	}
}

func kubernetesInstallBasePackages(t *Target, data interface{}) error {
	kubernetesBaseOSConfiguration, ok := data.(deployments.KubernetesBaseOSConfiguration)
	if !ok {
		return errors.New("couldn't access kubernetes base OS configuration")
	}

	t.ssh("zypper", "--non-interactive", "install", "--force",
		fmt.Sprintf("kubernetes-kubeadm=%s", kubernetesBaseOSConfiguration.KubeadmVersion),
		fmt.Sprintf("cri-o=%s", kubernetesBaseOSConfiguration.KubernetesVersion),
		fmt.Sprintf("kubernetes-client=%s", kubernetesBaseOSConfiguration.KubernetesVersion),
		fmt.Sprintf("kubernetes-common=%s", kubernetesBaseOSConfiguration.KubernetesVersion),
		fmt.Sprintf("kubernetes-kubelet=%s", kubernetesBaseOSConfiguration.KubernetesVersion),
	)
	// FIXME: do not ignore error, beta3 to beta4 package conflicts
	return nil
}

func kubernetesRestartServices(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "restart", "crio", "kubelet")
	return err
}
