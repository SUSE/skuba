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

package ssh

import (
	"path/filepath"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
)

type KubernetesUploadSecretsErrorBehavior uint

const (
	KubernetesUploadSecretsFailOnError     KubernetesUploadSecretsErrorBehavior = iota
	KubernetesUploadSecretsContinueOnError KubernetesUploadSecretsErrorBehavior = iota
)

func init() {
	stateMap["kubernetes.bootstrap.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsContinueOnError)
	stateMap["kubernetes.join.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsFailOnError)
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
