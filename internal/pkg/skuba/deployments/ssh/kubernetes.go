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
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

type KubernetesUploadSecretsErrorBehavior uint

const (
	KubernetesUploadSecretsFailOnError     KubernetesUploadSecretsErrorBehavior = iota
	KubernetesUploadSecretsContinueOnError KubernetesUploadSecretsErrorBehavior = iota
	kubeletTimeOutWait                                                          = 300
)

func init() {
	stateMap["kubernetes.bootstrap.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsContinueOnError)
	stateMap["kubernetes.join.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsFailOnError)
	stateMap["kubernetes.install-node-pattern"] = kubernetesInstallNodePattern
	stateMap["kubernetes.restart-services"] = kubernetesRestartServices
	stateMap["kubernetes.wait-for-kubelet"] = kubernetesWaitForKubelet
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

func kubernetesInstallNodePattern(t *Target, data interface{}) error {
	kubernetesBaseOSConfiguration, ok := data.(deployments.KubernetesBaseOSConfiguration)
	if !ok {
		return errors.New("couldn't access kubernetes base OS configuration")
	}

	v, err := version.ParseSemantic(kubernetesBaseOSConfiguration.CurrentVersion)
	if err != nil {
		return err
	}
	currentVersion := kubernetes.MajorMinorVersion(v)
	patternName := fmt.Sprintf("patterns-caasp-Node-%s", currentVersion)

	if kubernetesBaseOSConfiguration.UpdatedVersion != "" {
		updatedVersion := kubernetes.MajorMinorVersion(version.MustParseSemantic(kubernetesBaseOSConfiguration.UpdatedVersion))
		patternName = fmt.Sprintf("patterns-caasp-Node-%s-%s", currentVersion, updatedVersion)
	}
	_, _, err = t.ssh("zypper", "--non-interactive", "install", "--recommends", patternName)
	return err
}

func kubernetesRestartServices(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "restart", "crio", "kubelet")
	return err
}

func kubernetesWaitForKubelet(t *Target, data interface{}) error {
	clientSet, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "Error getting client set")
	}
	for i := 0; i < kubeletTimeOutWait; i++ {
		_, err := clientSet.CoreV1().Nodes().Get(t.target.Nodename, metav1.GetOptions{})
		if err != nil {
			klog.V(1).Infof("Still waiting for %s node to be registered, continuing...", t.target.Nodename)
		} else {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return errors.Wrap(err, fmt.Sprintf("Timed out waiting for %s", t.target.Nodename))
}
