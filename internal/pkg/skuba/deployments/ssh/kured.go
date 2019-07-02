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
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	stateMap["kured.deploy"] = kuredDeploy
}

func kuredDeploy(t *Target, data interface{}) error {
	kuredFiles, err := ioutil.ReadDir(skuba.KuredDir())
	if err != nil {
		return errors.Wrap(err, "could not read local kured directory")
	}

	defer t.ssh("rm -rf /tmp/kured.d")

	for _, f := range kuredFiles {
		if err := t.target.UploadFile(filepath.Join(skuba.KuredDir(), f.Name()), filepath.Join("/tmp/kured.d", f.Name())); err != nil {
			return err
		}
	}

	_, _, err = t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/kured.d")
	return err
}
