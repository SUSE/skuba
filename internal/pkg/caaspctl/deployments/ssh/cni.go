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
	"io/ioutil"
	"path"

	"github.com/pkg/errors"

	"github.com/SUSE/caaspctl/pkg/caaspctl"
)

func init() {
	stateMap["cni.deploy"] = cniDeploy
}

func cniDeploy(t *Target, data interface{}) error {
	cniFiles, err := ioutil.ReadDir(caaspctl.CniDir())
	if err != nil {
		return errors.Wrap(err, "could not read local cni directory")
	}

	defer t.ssh("rm -rf /tmp/cni.d")

	for _, f := range cniFiles {
		if err := t.target.UploadFile(path.Join(caaspctl.CniDir(), f.Name()), path.Join("/tmp/cni.d", f.Name())); err != nil {
			return err
		}
	}

	_, _, err = t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/cni.d")
	return err
}
