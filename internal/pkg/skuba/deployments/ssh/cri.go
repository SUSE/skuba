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
	"os"
	"path/filepath"

	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/pkg/errors"
)

func init() {
	stateMap["cri.configure"] = criConfigure
	stateMap["cri.sysconfig"] = criSysconfig
	stateMap["cri.start"] = criStart
}

func criConfigure(t *Target, data interface{}) error {
	criFiles, err := ioutil.ReadDir(skuba.CriConfDir())
	if err != nil {
		return errors.Wrap(err, "Could not read local cri directory: "+skuba.CriConfDir())
	}
	defer func() {
		_, _, err := t.ssh("rm -rf /tmp/crio.conf.d")
		if err != nil {
			// If the deferred function has any return values, they are discarded when the function completes
			// https://golang.org/ref/spec#Defer_statements
			fmt.Println("Could not delete the path /tmp/crio.conf.d")
		}
	}()

	for _, f := range criFiles {
		if err := t.target.UploadFile(filepath.Join(skuba.CriConfDir(), f.Name()), filepath.Join("/tmp/crio.conf.d", f.Name())); err != nil {
			return err
		}
	}

	if _, _, err = t.ssh("mkdir -p /etc/crio/crio.conf.d"); err != nil {
		return err
	}
	if _, _, err = t.ssh("cp -r /tmp/crio.conf.d/*.conf /etc/crio/crio.conf.d"); err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(skuba.ContainersDir(), "registries.conf")); err != nil {
		return nil
	}
	defer func() {
		_, _, err := t.ssh("rm -rf /tmp/containers")
		if err != nil {
			// If the deferred function has any return values, they are discarded when the function completes
			// https://golang.org/ref/spec#Defer_statements
			fmt.Println("Could not delete the path /tmp/containers")
		}
	}()

	if err := t.target.UploadFile(filepath.Join(skuba.ContainersDir(), "registries.conf"), filepath.Join("/tmp/containers", "registries.conf")); err != nil {
		return err
	}

	if _, _, err = t.ssh("mkdir -p /etc/containers"); err != nil {
		return err
	}
	_, _, err = t.ssh("cp -r /tmp/containers/*.conf /etc/containers")
	return err
}

// criSysconfig will enforce the package sysconfig configuration.
func criSysconfig(t *Target, data interface{}) error {
	_, _, err := t.ssh("cp -f /usr/share/fillup-templates/sysconfig.crio /etc/sysconfig/crio")
	return err
}

func criStart(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "enable", "--now", "crio")
	return err
}
