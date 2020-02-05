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

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
	"github.com/SUSE/skuba/pkg/skuba/actions/node/join"
)

func init() {
	stateMap["kubeadm.init"] = kubeadmInit
	stateMap["kubeadm.join"] = kubeadmJoin
	stateMap["kubeadm.reset"] = kubeadmReset
	stateMap["kubeadm.upgrade.apply"] = kubeadmUpgradeApply
	stateMap["kubeadm.upgrade.node"] = kubeadmUpgradeNode
}

var remoteKubeadmInitConfFile = filepath.Join("/tmp/", skubaconstants.KubeadmInitConfFile())

func kubeadmInit(t *Target, data interface{}) error {
	bootstrapConfiguration, ok := data.(deployments.BootstrapConfiguration)
	if !ok {
		return errors.New("couldn't access bootstrap configuration")
	}

	if err := t.target.UploadFile(skubaconstants.KubeadmInitConfFile(), remoteKubeadmInitConfFile); err != nil {
		return err
	}
	defer func() {
		_, _, err := t.ssh("rm", remoteKubeadmInitConfFile)
		if err != nil {
			// If the deferred function has any return values, they are discarded when the function completes
			// https://golang.org/ref/spec#Defer_statements
			fmt.Println("Could not delete the kubeadm init config file")
		}
	}()

	ignorePreflightErrors := ""
	ignorePreflightErrorsVal := bootstrapConfiguration.KubeadmExtraArgs["ignore-preflight-errors"]
	if len(ignorePreflightErrorsVal) > 0 {
		ignorePreflightErrors = "--ignore-preflight-errors=" + ignorePreflightErrorsVal
	}
	_, _, err := t.ssh("kubeadm", "init", "--config", remoteKubeadmInitConfFile, "--skip-token-print", ignorePreflightErrors, "-v", t.verboseLevel)
	return err
}

func kubeadmJoin(t *Target, data interface{}) error {
	api, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "could not retrieve the clientset from kubernetes")
	}
	joinConfiguration, ok := data.(deployments.JoinConfiguration)
	if !ok {
		return errors.New("couldn't access join configuration")
	}

	configPath, err := join.ConfigPath(api, joinConfiguration.Role, t.target)
	if err != nil {
		return errors.Wrap(err, "unable to configure path")
	}

	if err := t.target.UploadFile(configPath, remoteKubeadmInitConfFile); err != nil {
		return err
	}
	defer func() {
		_, _, err := t.ssh("rm", remoteKubeadmInitConfFile)
		if err != nil {
			// If the deferred function has any return values, they are discarded when the function completes
			// https://golang.org/ref/spec#Defer_statements
			fmt.Println("Could not delete the kubeadm init config file")
		}
	}()

	ignorePreflightErrors := ""
	ignorePreflightErrorsVal := joinConfiguration.KubeadmExtraArgs["ignore-preflight-errors"]
	if len(ignorePreflightErrorsVal) > 0 {
		ignorePreflightErrors = "--ignore-preflight-errors=" + ignorePreflightErrorsVal
	}
	_, _, err = t.ssh("kubeadm", "join", "--config", remoteKubeadmInitConfFile, ignorePreflightErrors, "-v", t.verboseLevel)
	return err
}

func kubeadmReset(t *Target, data interface{}) error {
	_, _, err := t.ssh("kubeadm", "reset", "--cri-socket", "/var/run/crio/crio.sock", "--ignore-preflight-errors", "all", "--force", "-v", t.verboseLevel)
	return err
}

func kubeadmUpgradeApply(t *Target, data interface{}) error {
	upgradeConfiguration, ok := data.(deployments.UpgradeConfiguration)
	if !ok {
		return errors.New("couldn't access upgrade configuration")
	}

	remoteKubeadmUpgradeConfFile := filepath.Join("/tmp/", skubaconstants.KubeadmUpgradeConfFile())
	if err := t.target.UploadFileContents(remoteKubeadmUpgradeConfFile, upgradeConfiguration.KubeadmConfigContents); err != nil {
		return err
	}
	defer func() {
		_, _, err := t.ssh("rm", remoteKubeadmUpgradeConfFile)
		if err != nil {
			// If the deferred function has any return values, they are discarded when the function completes
			// https://golang.org/ref/spec#Defer_statements
			fmt.Println("Could not delete the kubeadm upgrade config file")
		}
	}()

	_, _, err := t.ssh("kubeadm", "upgrade", "apply", "--config", remoteKubeadmUpgradeConfFile, "-y", "-v", t.verboseLevel)
	return err
}

func kubeadmUpgradeNode(t *Target, data interface{}) error {
	_, _, err := t.ssh("kubeadm", "upgrade", "node", "-v", t.verboseLevel)
	return err
}
