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

package kured

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"k8s.io/kubernetes/cmd/kubeadm/app/images"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"

	"github.com/pkg/errors"
)

type kuredConfiguration struct {
	KuredImage string
}

func renderKuredTemplate(kuredConfig kuredConfiguration, file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Errorf("could not create file %s", file)
	}

	template, err := template.New("").Parse(string(content))
	if err != nil {
		return errors.Errorf("could not parse template")
	}

	var rendered bytes.Buffer
	if err := template.Execute(&rendered, kuredConfig); err != nil {
		return errors.Errorf("could not render configuration")
	}

	if err := ioutil.WriteFile(file, rendered.Bytes(), 0644); err != nil {
		return errors.Errorf("could not write to %s: %s", file, err)
	}

	return nil
}

func FillKuredManifestFile() error {
	kuredImage := images.GetGenericImage(skuba.ImageRepository, "kured",
		kubernetes.CurrentAddonVersion(kubernetes.Kured))
	kuredConfig := kuredConfiguration{
		KuredImage: kuredImage,
	}

	return renderKuredTemplate(kuredConfig, filepath.Join("addons", "kured", "kured.yaml"))
}
