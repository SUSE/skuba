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

package cert

import (
	"os"

	"github.com/pkg/errors"

	"github.com/SUSE/skuba/internal/pkg/skuba/node"
	"github.com/SUSE/skuba/internal/pkg/skuba/oidc"
	"github.com/SUSE/skuba/pkg/skuba"
)

// GenerateCSRAndKey generates in-cluster services CSR and key into pki folder
func GenerateCSRAndKey() error {
	// load kubeadm-init.conf
	initCfg, err := node.LoadInitConfigurationFromFile(skuba.KubeadmInitConfFile())
	if err != nil {
		return err
	}

	// create 'pki' folder if not present
	f, err := os.Stat(skuba.PkiDir())
	if os.IsNotExist(err) || !f.IsDir() {
		// create pki folder
		if err := os.Mkdir(skuba.PkiDir(), 0700); err != nil {
			return errors.Wrapf(err, "unable to create directory: %s", skuba.PkiDir())
		}
	}

	// generate OIDC dex server CSR and key
	if err := oidc.GenerateServerCSRAndKey(oidc.DexCertCN, initCfg.APIServer.CertSANs, oidc.DexServerCertAndKeyBaseFileName); err != nil {
		return nil
	}

	// generate OIDC gangway server CSR and key
	if err := oidc.GenerateServerCSRAndKey(oidc.GangwayCertCN, initCfg.APIServer.CertSANs, oidc.GangwayServerCertAndKeyBaseFileName); err != nil {
		return nil
	}

	return nil
}
