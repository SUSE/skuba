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
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/SUSE/skuba/pkg/skuba/actions/node/join"
)

func init() {
	stateMap["kubeadm.init"] = kubeadmInit
	stateMap["kubeadm.join"] = kubeadmJoin
	stateMap["kubeadm.reset"] = kubeadmReset
	stateMap["kubeadm.upgrade.apply"] = kubeadmUpgradeApply
	stateMap["kubeadm.upgrade.node"] = kubeadmUpgradeNode
}

var remoteKubeadmInitConfFile = filepath.Join("/tmp/", skuba.KubeadmInitConfFile())

func kubeadmInit(t *Target, data interface{}) error {
	bootstrapConfiguration, ok := data.(deployments.BootstrapConfiguration)
	if !ok {
		return errors.New("couldn't access bootstrap configuration")
	}

	if err := t.target.UploadFile(skuba.KubeadmInitConfFile(), remoteKubeadmInitConfFile); err != nil {
		return err
	}
	defer t.ssh("rm", remoteKubeadmInitConfFile)

	ignorePreflightErrors := ""
	ignorePreflightErrorsVal := bootstrapConfiguration.KubeadmExtraArgs["ignore-preflight-errors"]
	if len(ignorePreflightErrorsVal) > 0 {
		ignorePreflightErrors = "--ignore-preflight-errors=" + ignorePreflightErrorsVal
	}
	_, _, err := t.ssh("kubeadm", "init", "--config", remoteKubeadmInitConfFile, "--skip-token-print", ignorePreflightErrors)
	return err
}

func kubeadmJoin(t *Target, data interface{}) error {
	joinConfiguration, ok := data.(deployments.JoinConfiguration)
	if !ok {
		return errors.New("couldn't access join configuration")
	}

	configPath, err := join.ConfigPath(joinConfiguration.Role, t.target)
	if err != nil {
		return errors.Wrap(err, "unable to configure path")
	}

	if err := t.target.UploadFile(configPath, remoteKubeadmInitConfFile); err != nil {
		return err
	}
	defer t.ssh("rm", remoteKubeadmInitConfFile)

	ignorePreflightErrors := ""
	ignorePreflightErrorsVal := joinConfiguration.KubeadmExtraArgs["ignore-preflight-errors"]
	if len(ignorePreflightErrorsVal) > 0 {
		ignorePreflightErrors = "--ignore-preflight-errors=" + ignorePreflightErrorsVal
	}
	_, _, err = t.ssh("kubeadm", "join", "--config", remoteKubeadmInitConfFile, ignorePreflightErrors)
	return err
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

func kubeadmUpgradeApply(t *Target, data interface{}) error {
	upgradeConfiguration, ok := data.(deployments.UpgradeConfiguration)
	if !ok {
		return errors.New("couldn't access upgrade configuration")
	}

	remoteKubeadmUpgradeConfFile := filepath.Join("/tmp/", skuba.KubeadmUpgradeConfFile())
	if err := t.target.UploadFileContents(remoteKubeadmUpgradeConfFile, upgradeConfiguration.KubeadmConfigContents); err != nil {
		return err
	}
	defer t.ssh("rm", remoteKubeadmUpgradeConfFile)

	_, _, err := t.ssh("kubeadm", "upgrade", "apply", "--config", remoteKubeadmUpgradeConfFile, "-y")
	return err
}

func kubeadmUpgradeNode(t *Target, data interface{}) error {
	_, _, err := t.ssh("kubeadm", "upgrade", "node")
	return err
}
