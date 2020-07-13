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
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"github.com/SUSE/skuba/internal/pkg/skuba/addons"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/util"
	"github.com/SUSE/skuba/pkg/skuba"
)

// Basic initial cluster configuration
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
	// Note: UseHyperKube can be removed when we drop the support of
	// provisioning clusters of version 1.17.
	UseHyperKube bool
	CniPlugin    kubernetes.Addon
}

func (initConfiguration InitConfiguration) ControlPlaneHost() string {
	return util.ControlPlaneHost(initConfiguration.ControlPlane)
}

func (initConfiguration InitConfiguration) ControlPlaneHostAndPort() string {
	return util.ControlPlaneHostAndPort(initConfiguration.ControlPlane)
}

func NewInitConfiguration(clusterName, cloudProvider, controlPlane, kubernetesDesiredVersion string, strictCapDefaults bool, cniPlugin string) (InitConfiguration, error) {
	kubernetesVersion := kubernetes.LatestVersion()
	var err error
	needsHyperKube := false
	if kubernetesDesiredVersion != "" {
		kubernetesVersion, err = versionutil.ParseSemantic(kubernetesDesiredVersion)
		if err != nil || !kubernetes.IsVersionAvailable(kubernetesVersion) {
			return InitConfiguration{}, fmt.Errorf("Version %s does not exist or cannot be parsed.\n", kubernetesDesiredVersion)
		}
	}

	// Without this, it will be impossible to greenfield an older caasp cluster:
	// defaults have been changed in 1.17, so we *need* to have UseHyperKubeImage: set into the init configuration.
	if kubernetesVersion.Minor() < 18 {
		needsHyperKube = true
	}

	return InitConfiguration{
		ClusterName:       clusterName,
		CloudProvider:     cloudProvider,
		ControlPlane:      controlPlane,
		PauseImage:        kubernetes.ComponentContainerImageForClusterVersion(kubernetes.Pause, kubernetesVersion),
		KubernetesVersion: kubernetesVersion,
		ImageRepository:   skuba.ImageRepository,
		EtcdImageTag:      kubernetes.ComponentVersionForClusterVersion(kubernetes.Etcd, kubernetesVersion),
		CoreDNSImageTag:   kubernetes.ComponentVersionForClusterVersion(kubernetes.CoreDNS, kubernetesVersion),
		StrictCapDefaults: strictCapDefaults,
		UseHyperKube:      needsHyperKube,
		CniPlugin:         kubernetes.Addon(cniPlugin),
	}, nil
}

// Init creates a cluster definition scaffold in the local machine, in the current
// folder, at a directory named after ClusterName provided in the InitConfiguration
// parameter
func Init(initConfiguration InitConfiguration) error {
	if _, err := os.Stat(initConfiguration.ClusterName); err == nil {
		return errors.Errorf("cluster configuration directory %q already exists", initConfiguration.ClusterName)
	}
	if addon, found := addons.Addons[initConfiguration.CniPlugin]; !found || addon.AddOnType != addons.CniAddOn {
		return fmt.Errorf("unknown CNI plugin provided: %s", initConfiguration.CniPlugin)
	}

	// write configuration files
	if err := writeScaffoldFiles(initConfiguration); err != nil {
		return err
	}
	if err := writeKubeadmFiles(initConfiguration); err != nil {
		return err
	}
	if err := writeAddonConfigFiles(initConfiguration); err != nil {
		return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("[init] configuration files written, unable to get directory")
		return nil
	}

	fmt.Printf("[init] configuration files written to %s\n", currentDir)
	return nil
}

func isAddonRequired(addon addons.Addon, initConfiguration InitConfiguration) bool {
	if !addon.IsPresentForClusterVersion(initConfiguration.KubernetesVersion) {
		return false
	}
	if addon.AddOnType == addons.CniAddOn && initConfiguration.CniPlugin != addon.Addon {
		return false
	}
	return true
}

func writeScaffoldFiles(initConfiguration InitConfiguration) error {
	scaffoldFilesToWrite := criScaffoldFiles["criconfig"]
	kubernetesVersion := initConfiguration.KubernetesVersion
	if kubernetesVersion.Minor() < 18 {
		scaffoldFilesToWrite = criScaffoldFiles["sysconfig"]
	}

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
		_, err = f.WriteString(str)
		if err != nil {
			return errors.Wrapf(err, "unable to write template to file %s", f.Name())
		}
		if err := f.Chmod(0600); err != nil {
			return errors.Wrapf(err, "unable to chmod file %s", f.Name())
		}
		if err := f.Close(); err != nil {
			return errors.Wrapf(err, "unable to close file %s", f.Name())
		}
	}
	return nil
}

func writeKubeadmFiles(initConfiguration InitConfiguration) error {
	// Write kubeadm-init.conf and kubeadm-join.conf.d templates
	if err := writeKubeadmInitConf(initConfiguration); err != nil {
		return err
	}
	if err := os.MkdirAll(skuba.JoinConfDir(), 0700); err != nil {
		return errors.Wrapf(err, "could not create directory %q", skuba.JoinConfDir())
	}
	if err := writeKubeadmJoinMasterConf(initConfiguration); err != nil {
		return err
	}
	if err := writeKubeadmJoinWorkerConf(initConfiguration); err != nil {
		return err
	}
	return nil
}

func writeAddonConfigFiles(initConfiguration InitConfiguration) error {
	// Write addon configuration files
	addonConfiguration := addons.AddonConfiguration{
		ClusterVersion: initConfiguration.KubernetesVersion,
		ControlPlane:   initConfiguration.ControlPlane,
		ClusterName:    initConfiguration.ClusterName,
	}
	for addonName, addon := range addons.Addons {
		if !isAddonRequired(addon, initConfiguration) {
			continue
		}
		if err := addon.Write(addonConfiguration); err != nil {
			return errors.Wrapf(err, "could not write %q addon configuration", addonName)
		}
	}
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

func writeKubeadmInitConf(initConfiguration InitConfiguration) error {
	initCfg := kubeadmapi.InitConfiguration{
		ClusterConfiguration: kubeadmapi.ClusterConfiguration{
			APIServer: kubeadmapi.APIServer{
				CertSANs: []string{initConfiguration.ControlPlaneHost()},
				ControlPlaneComponent: kubeadmapi.ControlPlaneComponent{
					ExtraArgs: map[string]string{
						"oidc-issuer-url":                  fmt.Sprintf("https://%s:32000", initConfiguration.ControlPlaneHost()),
						"oidc-client-id":                   "oidc",
						"oidc-ca-file":                     "/etc/kubernetes/pki/ca.crt",
						"oidc-username-claim":              "email",
						"oidc-groups-claim":                "groups",
						"service-account-issuer":           "kubernetes.default.svc",
						"service-account-signing-key-file": "/etc/kubernetes/pki/sa.key",
					},
				},
			},
			ClusterName:          initConfiguration.ClusterName,
			ControlPlaneEndpoint: initConfiguration.ControlPlaneHostAndPort(),
			DNS: kubeadmapi.DNS{
				Type: kubeadmapi.CoreDNS,
				ImageMeta: kubeadmapi.ImageMeta{
					ImageRepository: initConfiguration.ImageRepository,
					ImageTag:        initConfiguration.CoreDNSImageTag,
				},
			},
			Etcd: kubeadmapi.Etcd{
				Local: &kubeadmapi.LocalEtcd{
					ImageMeta: kubeadmapi.ImageMeta{
						ImageRepository: initConfiguration.ImageRepository,
						ImageTag:        initConfiguration.EtcdImageTag,
					},
				},
			},
			ImageRepository:   initConfiguration.ImageRepository,
			KubernetesVersion: initConfiguration.KubernetesVersion.String(),
			Networking: kubeadmapi.Networking{
				PodSubnet:     "10.244.0.0/16",
				ServiceSubnet: "10.96.0.0/12",
			},
			UseHyperKubeImage: initConfiguration.UseHyperKube,
		},
	}
	if len(initConfiguration.CloudProvider) > 0 {
		updateInitConfigurationWithCloudIntegration(&initCfg, initConfiguration)
	}
	kubeadm.UpdateClusterConfigurationWithClusterVersion(&initCfg, initConfiguration.KubernetesVersion)
	initCfgContents, err := kubeadmconfigutil.MarshalInitConfigurationToBytes(&initCfg, schema.GroupVersion{
		Group:   "kubeadm.k8s.io",
		Version: kubeadm.GetKubeadmApisVersion(initConfiguration.KubernetesVersion),
	})
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(skuba.KubeadmInitConfFile(), initCfgContents, 0600); err != nil {
		return errors.Wrap(err, "error writing init configuration")
	}
	return nil
}

func writeKubeadmJoinMasterConf(initConfiguration InitConfiguration) error {
	joinCfg := kubeadmapi.JoinConfiguration{
		Discovery: kubeadmapi.Discovery{
			BootstrapToken: &kubeadmapi.BootstrapTokenDiscovery{
				APIServerEndpoint:        initConfiguration.ControlPlaneHostAndPort(),
				UnsafeSkipCAVerification: true,
			},
		},
		ControlPlane: &kubeadmapi.JoinControlPlane{},
	}
	if len(initConfiguration.CloudProvider) > 0 {
		updateJoinConfigurationWithCloudIntegration(&joinCfg, initConfiguration)
	}
	joinCfgContents, err := kubeadmutil.MarshalToYamlForCodecs(&joinCfg, schema.GroupVersion{
		Group:   "kubeadm.k8s.io",
		Version: kubeadm.GetKubeadmApisVersion(initConfiguration.KubernetesVersion),
	}, kubeadmscheme.Codecs)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(skuba.MasterConfTemplateFile(), joinCfgContents, 0600); err != nil {
		return errors.Wrap(err, "error writing control plane join configuration")
	}
	return nil
}

func writeKubeadmJoinWorkerConf(initConfiguration InitConfiguration) error {
	joinCfg := kubeadmapi.JoinConfiguration{
		Discovery: kubeadmapi.Discovery{
			BootstrapToken: &kubeadmapi.BootstrapTokenDiscovery{
				APIServerEndpoint:        initConfiguration.ControlPlaneHostAndPort(),
				UnsafeSkipCAVerification: true,
			},
		},
	}
	if len(initConfiguration.CloudProvider) > 0 {
		updateJoinConfigurationWithCloudIntegration(&joinCfg, initConfiguration)
	}
	joinCfgContents, err := kubeadmutil.MarshalToYamlForCodecs(&joinCfg, schema.GroupVersion{
		Group:   "kubeadm.k8s.io",
		Version: kubeadm.GetKubeadmApisVersion(initConfiguration.KubernetesVersion),
	}, kubeadmscheme.Codecs)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(skuba.WorkerConfTemplateFile(), joinCfgContents, 0600); err != nil {
		return errors.Wrap(err, "error writing worker join configuration")
	}
	return nil
}

func updateInitConfigurationWithCloudIntegration(initCfg *kubeadmapi.InitConfiguration, initConfiguration InitConfiguration) {
	if initCfg.APIServer.ExtraArgs == nil {
		initCfg.APIServer.ExtraArgs = map[string]string{}
	}
	initCfg.APIServer.ExtraArgs["cloud-provider"] = initConfiguration.CloudProvider
	if initCfg.ControllerManager.ExtraArgs == nil {
		initCfg.ControllerManager.ExtraArgs = map[string]string{}
	}
	initCfg.ControllerManager.ExtraArgs["cloud-provider"] = initConfiguration.CloudProvider
	if initCfg.NodeRegistration.KubeletExtraArgs == nil {
		initCfg.NodeRegistration.KubeletExtraArgs = map[string]string{}
	}
	initCfg.NodeRegistration.KubeletExtraArgs["cloud-provider"] = initConfiguration.CloudProvider

	switch initConfiguration.CloudProvider {
	case "aws":
		initCfg.ControllerManager.ExtraArgs["allocate-node-cidrs"] = "false"
	case "azure":
		initCfg.APIServer.ExtraArgs["cloud-config"] = skuba.AzureConfigRuntimeFile()
		initCfg.APIServer.ExtraVolumes = append(initCfg.APIServer.ExtraVolumes, kubeadmapi.HostPathMount{
			Name:      "cloud-config",
			HostPath:  skuba.AzureConfigRuntimeFile(),
			MountPath: skuba.AzureConfigRuntimeFile(),
			ReadOnly:  true,
			PathType:  v1.HostPathFileOrCreate,
		})
		initCfg.ControllerManager.ExtraArgs["cloud-config"] = skuba.AzureConfigRuntimeFile()
		initCfg.ControllerManager.ExtraVolumes = append(initCfg.ControllerManager.ExtraVolumes, kubeadmapi.HostPathMount{
			Name:      "cloud-config",
			HostPath:  skuba.AzureConfigRuntimeFile(),
			MountPath: skuba.AzureConfigRuntimeFile(),
			ReadOnly:  true,
			PathType:  v1.HostPathFileOrCreate,
		})
		initCfg.NodeRegistration.KubeletExtraArgs["cloud-config"] = skuba.AzureConfigRuntimeFile()
	case "openstack":
		initCfg.APIServer.ExtraArgs["cloud-config"] = skuba.OpenstackConfigRuntimeFile()
		initCfg.APIServer.ExtraVolumes = append(initCfg.APIServer.ExtraVolumes, kubeadmapi.HostPathMount{
			Name:      "cloud-config",
			HostPath:  skuba.OpenstackConfigRuntimeFile(),
			MountPath: skuba.OpenstackConfigRuntimeFile(),
			ReadOnly:  true,
			PathType:  v1.HostPathFileOrCreate,
		})
		initCfg.ControllerManager.ExtraArgs["cloud-config"] = skuba.OpenstackConfigRuntimeFile()
		initCfg.ControllerManager.ExtraVolumes = append(initCfg.ControllerManager.ExtraVolumes, kubeadmapi.HostPathMount{
			Name:      "cloud-config",
			HostPath:  skuba.OpenstackConfigRuntimeFile(),
			MountPath: skuba.OpenstackConfigRuntimeFile(),
			ReadOnly:  true,
			PathType:  v1.HostPathFileOrCreate,
		})
		initCfg.NodeRegistration.KubeletExtraArgs["cloud-config"] = skuba.OpenstackConfigRuntimeFile()
	case "vsphere":
		initCfg.APIServer.ExtraArgs["cloud-config"] = skuba.VSphereConfigRuntimeFile()
		initCfg.APIServer.ExtraVolumes = append(initCfg.APIServer.ExtraVolumes, kubeadmapi.HostPathMount{
			Name:      "cloud-config",
			HostPath:  skuba.VSphereConfigRuntimeFile(),
			MountPath: skuba.VSphereConfigRuntimeFile(),
			ReadOnly:  true,
			PathType:  v1.HostPathFileOrCreate,
		})
		initCfg.ControllerManager.ExtraArgs["cloud-config"] = skuba.VSphereConfigRuntimeFile()
		initCfg.ControllerManager.ExtraVolumes = append(initCfg.ControllerManager.ExtraVolumes, kubeadmapi.HostPathMount{
			Name:      "cloud-config",
			HostPath:  skuba.VSphereConfigRuntimeFile(),
			MountPath: skuba.VSphereConfigRuntimeFile(),
			ReadOnly:  true,
			PathType:  v1.HostPathFileOrCreate,
		})
		initCfg.NodeRegistration.KubeletExtraArgs["cloud-config"] = skuba.VSphereConfigRuntimeFile()
	}
}

func updateJoinConfigurationWithCloudIntegration(joinCfg *kubeadmapi.JoinConfiguration, initConfiguration InitConfiguration) {
	if joinCfg.NodeRegistration.KubeletExtraArgs == nil {
		joinCfg.NodeRegistration.KubeletExtraArgs = map[string]string{}
	}
	joinCfg.NodeRegistration.KubeletExtraArgs["cloud-provider"] = initConfiguration.CloudProvider

	switch initConfiguration.CloudProvider {
	case "azure":
		joinCfg.NodeRegistration.KubeletExtraArgs["cloud-config"] = skuba.AzureConfigRuntimeFile()
	case "openstack":
		joinCfg.NodeRegistration.KubeletExtraArgs["cloud-config"] = skuba.OpenstackConfigRuntimeFile()
	case "vsphere":
		joinCfg.NodeRegistration.KubeletExtraArgs["cloud-config"] = skuba.VSphereConfigRuntimeFile()
	}
}
