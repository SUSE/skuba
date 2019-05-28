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
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	stateMap["psp.deploy"] = pspDeploy
}

func pspDeploy(t *Target, data interface{}) error {
	pspFiles, err := ioutil.ReadDir(skuba.PspDir())
	if err != nil {
		return errors.Wrap(err, "could not read local psp directory")
	}

	defer t.ssh("rm -rf /tmp/psp.d")

	for _, f := range pspFiles {
		if err := t.target.UploadFile(filepath.Join(skuba.PspDir(), f.Name()), filepath.Join("/tmp/psp.d", f.Name())); err != nil {
			return err
		}
	}

	_, _, err = t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/psp.d")
	return err

}
