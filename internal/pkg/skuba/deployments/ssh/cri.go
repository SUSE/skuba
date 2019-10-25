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
	"io/ioutil"
	"path/filepath"

	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/pkg/errors"
)

func init() {
	stateMap["cri.configure"] = criConfigure
	stateMap["cri.start"] = criStart
}

func criConfigure(t *Target, data interface{}) error {
	criFiles, err := ioutil.ReadDir(skuba.CriDir())
	if err != nil {
		return errors.Wrap(err, "Could not read local cri directory: "+skuba.CriDir())
	}
	defer func() {
		_, _, err := t.ssh("rm -rf /tmp/cri.d")
		if err != nil {
			// If the deferred function has any return values, they are discarded when the function completes
			// https://golang.org/ref/spec#Defer_statements
			fmt.Println("Could not delete the cri.d config path")
		}
	}()

	for _, f := range criFiles {
		if err := t.target.UploadFile(filepath.Join(skuba.CriDir(), f.Name()), filepath.Join("/tmp/cri.d", f.Name())); err != nil {
			return err
		}
	}

	if _, _, err = t.ssh("mv -f /etc/sysconfig/crio /etc/sysconfig/crio.backup"); err != nil {
		return err
	}
	_, _, err = t.ssh("mv -f /tmp/cri.d/default_flags /etc/sysconfig/crio")
	return err
}

func criStart(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "enable", "--now", "crio")
	return err
}
