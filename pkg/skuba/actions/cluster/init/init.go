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

package cluster

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"

	"github.com/SUSE/skuba/internal/pkg/skuba/addons"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

type InitConfiguration struct {
	ClusterName       string
	ControlPlane      string
	PauseImage        string
	KubernetesVersion *versionutil.Version
	ImageRepository   string
	EtcdImageTag      string
	CoreDNSImageTag   string
	CloudProvider     string
	StrictCapDefaults bool
}

func (initConfiguration InitConfiguration) ControlPlaneHost() string {
	controlPlane, _, err := kubeadmutil.ParseHostPort(initConfiguration.ControlPlane)
	if err != nil {
		return ""
	}
	return controlPlane
}

func (initConfiguration InitConfiguration) ControlPlaneHostAndPort() string {
	controlPlaneHost, controlPlanePort, err := kubeadmutil.ParseHostPort(initConfiguration.ControlPlane)
	if err != nil {
		return ""
	}
	if controlPlanePort == "" {
		controlPlanePort = "6443"
	}
	return fmt.Sprintf("%s:%s", controlPlaneHost, controlPlanePort)
}

func NewInitConfiguration(clusterName, cloudProvider, controlPlane, kubernetesDesiredVersion string, strictCapDefaults bool) (InitConfiguration, error) {
	kubernetesVersion := kubernetes.LatestVersion()
	var err error
	if kubernetesDesiredVersion != "" {
		kubernetesVersion, err = versionutil.ParseSemantic(kubernetesDesiredVersion)
		if err != nil || !kubernetes.IsVersionAvailable(kubernetesVersion) {
			return InitConfiguration{}, fmt.Errorf("Version %s does not exist or cannot be parsed.\n", kubernetesDesiredVersion)
		}
	}

	return InitConfiguration{
		ClusterName:       clusterName,
		CloudProvider:     cloudProvider,
		ControlPlane:      controlPlane,
		PauseImage:        images.GetGenericImage(skuba.ImageRepository, "pause", kubernetes.ComponentVersionForClusterVersion(kubernetes.Pause, kubernetesVersion)),
		KubernetesVersion: kubernetesVersion,
		ImageRepository:   skuba.ImageRepository,
		EtcdImageTag:      kubernetes.ComponentVersionForClusterVersion(kubernetes.Etcd, kubernetesVersion),
		CoreDNSImageTag:   kubernetes.ComponentVersionForClusterVersion(kubernetes.CoreDNS, kubernetesVersion),
		StrictCapDefaults: strictCapDefaults,
	}, nil
}

// Init creates a cluster definition scaffold in the local machine, in the current
// folder, at a directory named after ClusterName provided in the InitConfiguration
// parameter
//
// FIXME: being this a part of the go API accept the toplevel directory instead of
//        using the PWD
func Init(initConfiguration InitConfiguration) error {
	if _, err := os.Stat(initConfiguration.ClusterName); err == nil {
		return errors.Errorf("cluster configuration directory %q already exists", initConfiguration.ClusterName)
	}

	scaffoldFilesToWrite := scaffoldFiles
	if len(initConfiguration.CloudProvider) > 0 {
		if cloudScaffoldFiles, found := cloudScaffoldFiles[initConfiguration.CloudProvider]; found {
			scaffoldFilesToWrite = append(scaffoldFilesToWrite, cloudScaffoldFiles...)
		} else {
			klog.Fatalf("unknown cloud provider integration provided: %s", initConfiguration.CloudProvider)
		}
	}

	if err := os.MkdirAll(initConfiguration.ClusterName, 0700); err != nil {
		return errors.Wrapf(err, "could not create cluster directory %q", initConfiguration.ClusterName)
	}
	if err := os.Chdir(initConfiguration.ClusterName); err != nil {
		return errors.Wrapf(err, "could not change to cluster directory %q", initConfiguration.ClusterName)
	}
	for _, file := range scaffoldFilesToWrite {
		filePath, _ := filepath.Split(file.Location)
		if filePath != "" {
			if err := os.MkdirAll(filePath, 0700); err != nil {
				return errors.Wrapf(err, "could not create directory %q", filePath)
			}
		}
		f, err := os.Create(file.Location)
		if err != nil {
			return errors.Wrapf(err, "could not create file %q", file.Location)
		}
		str, err := renderTemplate(file.Content, initConfiguration)
		if err != nil {
			return errors.Wrap(err, "unable to render template")
		}
		f.WriteString(str)
		f.Chmod(0600)
		f.Close()
	}

	addonConfiguration := addons.AddonConfiguration{
		ClusterVersion: initConfiguration.KubernetesVersion,
		ControlPlane:   initConfiguration.ControlPlane,
		ClusterName:    initConfiguration.ClusterName,
	}
	for addonName, addon := range addons.Addons {
		if !addon.IsPresentForClusterVersion(initConfiguration.KubernetesVersion) {
			continue
		}
		if err := addon.Write(addonConfiguration); err != nil {
			return errors.Wrapf(err, "could not write %q addon configuration", addonName)
		}
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("[init] configuration files written, unable to get directory")
		return nil
	}

	fmt.Printf("[init] configuration files written to %s\n", currentDir)
	return nil
}

func renderTemplate(templateContents string, initConfiguration InitConfiguration) (string, error) {
	template, err := template.New("").Parse(templateContents)
	if err != nil {
		return "", errors.Wrap(err, "could not parse template")
	}
	var rendered bytes.Buffer
	if err := template.Execute(&rendered, initConfiguration); err != nil {
		return "", errors.Wrap(err, "could not render configuration")
	}
	return rendered.String(), nil
}
