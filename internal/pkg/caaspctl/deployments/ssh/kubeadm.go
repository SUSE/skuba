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
	"github.com/pkg/errors"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments"
	"github.com/SUSE/caaspctl/pkg/caaspctl"
	node "github.com/SUSE/caaspctl/pkg/caaspctl/actions/node/join"
)

func init() {
	stateMap["kubeadm.init"] = kubeadmInit
	stateMap["kubeadm.join"] = kubeadmJoin
	stateMap["kubeadm.reset"] = kubeadmReset
}

func kubeadmReset(t *Target, data interface{}) error {
	resetConfiguration, ok := data.(deployments.ResetConfiguration)
	if !ok {
		return errors.New("couldn't access reset configuration")
	}

	ignorePreflightErrors := ""
	ignorePreflightErrorsVal := resetConfiguration.KubeadmExtraArgs["ignore-preflight-errors"]
	if len(ignorePreflightErrorsVal) > 0 {
		ignorePreflightErrors = "--ignore-preflight-errors=" + ignorePreflightErrorsVal
	}
	_, _, err := t.ssh("kubeadm", "reset", "--cri-socket", "/var/run/crio/crio.sock", "--force", ignorePreflightErrors)
	return err
}

func kubeadmInit(t *Target, data interface{}) error {
	bootstrapConfiguration, ok := data.(deployments.BootstrapConfiguration)
	if !ok {
		return errors.New("couldn't access bootstrap configuration")
	}

	if err := t.target.UploadFile(caaspctl.KubeadmInitConfFile(), "/tmp/kubeadm.conf"); err != nil {
		return err
	}
	defer t.ssh("rm", "/tmp/kubeadm.conf")

	ignorePreflightErrors := ""
	ignorePreflightErrorsVal := bootstrapConfiguration.KubeadmExtraArgs["ignore-preflight-errors"]
	if len(ignorePreflightErrorsVal) > 0 {
		ignorePreflightErrors = "--ignore-preflight-errors=" + ignorePreflightErrorsVal
	}
	_, _, err := t.ssh("kubeadm", "init", "--config", "/tmp/kubeadm.conf", "--skip-token-print", ignorePreflightErrors)
	return err
}

func kubeadmJoin(t *Target, data interface{}) error {
	joinConfiguration, ok := data.(deployments.JoinConfiguration)
	if !ok {
		return errors.New("couldn't access join configuration")
	}

	if err := t.target.UploadFile(node.ConfigPath(joinConfiguration.Role, t.target), "/tmp/kubeadm.conf"); err != nil {
		return err
	}
	defer t.ssh("rm", "/tmp/kubeadm.conf")

	ignorePreflightErrors := ""
	ignorePreflightErrorsVal := joinConfiguration.KubeadmExtraArgs["ignore-preflight-errors"]
	if len(ignorePreflightErrorsVal) > 0 {
		ignorePreflightErrors = "--ignore-preflight-errors=" + ignorePreflightErrorsVal
	}
	_, _, err := t.ssh("kubeadm", "join", "--config", "/tmp/kubeadm.conf", ignorePreflightErrors)
	return err
}
