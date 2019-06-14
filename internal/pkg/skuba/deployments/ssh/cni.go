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

	"github.com/SUSE/skuba/internal/pkg/skuba/cni"
	"github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	stateMap["cni.deploy"] = cniDeploy
	stateMap["cni.render"] = cniRender
	stateMap["cni.cilium-update-configmap"] = ciliumUpdateConfigMap
}

func cniRender(t *Target, data interface{}) error {
	if err := cni.FillCiliumManifestFile(); err != nil {
		return err
	}
	return nil
}

func cniDeploy(t *Target, data interface{}) error {
	if err := cni.CreateCiliumSecret(); err != nil {
		return errors.Wrap(err, "unable to create cilium secrets")
	}
	if err := cni.CreateOrUpdateCiliumConfigMap(); err != nil {
		return errors.Wrap(err, "unable to create or update cilium config map")
	}
	cniFiles, err := ioutil.ReadDir(skuba.CniDir())
	if err != nil {
		return errors.Wrap(err, "could not read local cni directory")
	}

	defer t.ssh("rm -rf /tmp/cni.d")

	for _, f := range cniFiles {
		if err := t.target.UploadFile(filepath.Join(skuba.CniDir(), f.Name()), filepath.Join("/tmp/cni.d", f.Name())); err != nil {
			return err
		}
	}

	_, _, err = t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/cni.d")
	return err
}

func ciliumUpdateConfigMap(t *Target, data interface{}) error {
	if err := cni.CreateOrUpdateCiliumConfigMap(); err != nil {
		return err
	}

	return cni.AnnotateCiliumDaemonsetWithCurrentTimestamp()
}
