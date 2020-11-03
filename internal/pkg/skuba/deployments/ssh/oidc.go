/*
 * Copyright (c) 2020 SUSE LLC.
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
	"os"
	"path/filepath"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/node"
	"github.com/SUSE/skuba/internal/pkg/skuba/oidc"
	"github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	stateMap["oidc.ca.upload"] = oidcCAUpload
}

// oidcCAUpload upload local OIDC CA certificate to remote server
// path specified in the local file kubeadm-init.conf key oidc-ca-file
func oidcCAUpload(t *Target, data interface{}) error {
	// check OIDC CA cert exist
	if certExist, _ := oidc.IsCACertAndKeyExist(); !certExist {
		return nil
	}
	// upload OIDC CA cert to control plane nodes only
	if *t.target.Role != deployments.MasterRole {
		return nil
	}

	// read kubeadm-init.conf and then uploads the oidc-ca to the file specified in `oidc-ca-file`
	remoteOIDCCAFilePath, err := getOIDCCAFile()
	if err != nil {
		return err
	}
	caCertPath := filepath.Join(skuba.PkiDir(), oidc.CACertFileName)
	f, err := os.Stat(caCertPath)
	if err != nil {
		return err
	}
	if err := t.target.UploadFile(filepath.Join(skuba.PkiDir(), oidc.CACertFileName), remoteOIDCCAFilePath, f.Mode()); err != nil {
		return err
	}
	return nil
}

// getOIDCCAFile returns local file kubeadm-init.conf key oidc-ca-file
func getOIDCCAFile() (string, error) {
	// load kubeadm-init.conf
	initCfg, err := node.LoadInitConfigurationFromFile(skuba.KubeadmInitConfFile())
	if err != nil {
		return "", err
	}
	return initCfg.APIServer.ExtraArgs["oidc-ca-file"], nil
}
