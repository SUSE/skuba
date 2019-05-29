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
