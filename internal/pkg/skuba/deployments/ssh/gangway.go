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

	"github.com/SUSE/skuba/internal/pkg/skuba/gangway"
	"github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	stateMap["gangway.deploy"] = gangwayDeploy
}

func gangwayDeploy(t *Target, data interface{}) error {
	if err := gangway.CreateGangwaySessionKey(); err != nil {
		return errors.Wrap(err, "unable to create gangway session key")
	}
	if err := gangway.CreateGangwayCert(); err != nil {
		return errors.Wrap(err, "unable to create gangway certificate")
	}

	gangwayFiles, err := ioutil.ReadDir(skuba.GangwayDir())
	if err != nil {
		return errors.Wrap(err, "could not read local gangway directory")
	}

	defer t.ssh("rm -rf /tmp/gangway.d")

	for _, f := range gangwayFiles {
		if err := t.target.UploadFile(filepath.Join(skuba.GangwayDir(), f.Name()), filepath.Join("/tmp/gangway.d", f.Name())); err != nil {
			return err
		}
	}

	_, _, err = t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/gangway.d")
	return err
}
