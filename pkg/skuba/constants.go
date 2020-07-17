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

package skuba

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/version"
	"path"
	"path/filepath"

	"k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
)

const (
	CRISocket  = "/var/run/crio/crio.sock"
	SUSECNIDir = "/usr/lib/cni"
	// MaxNodeNameLength is the maximum node name length accepted by kubelet.
	MaxNodeNameLength = 63
)

func KubeadmInitConfFile() string {
	return "kubeadm-init.conf"
}

func KubeadmUpgradeConfFile() string {
	return "kubeadm-upgrade.conf"
}

func JoinConfDir() string {
	return "kubeadm-join.conf.d"
}

func MasterConfTemplateFile() string {
	return filepath.Join(JoinConfDir(), "master.conf.template")
}

func WorkerConfTemplateFile() string {
	return filepath.Join(JoinConfDir(), "worker.conf.template")
}

func MachineConfFile(target string) string {
	return filepath.Join(JoinConfDir(), fmt.Sprintf("%s.conf", target))
}

func TemplatePathForRole(role deployments.Role) string {
	switch role {
	case deployments.MasterRole:
		return MasterConfTemplateFile()
	case deployments.WorkerRole:
		return WorkerConfTemplateFile()
	}
	return ""
}

func AddonsDir() string {
	return "addons"
}

func ContainersDir() string {
	return filepath.Join(AddonsDir(), "containers")
}

func CriDir() string {
	return filepath.Join(AddonsDir(), "cri")
}

func CriDockerDefaultsConfFile() string {
	return filepath.Join(CriDir(), "default_flags")
}

func CriConfDir() string {
	return filepath.Join(AddonsDir(), "cri/conf.d")
}

func CriDefaultsConfFile() string {
	return filepath.Join(CriConfDir(), "01-caasp.conf")
}

func CriConfFolderReadmeFile() string {
	return filepath.Join(CriConfDir(), "README")
}

func KubeConfigAdminFile() string {
	return "admin.conf"
}

func PkiDir() string {
	return "pki"
}

// CloudDir returns the reletive location for cloud config files
func CloudDir() string {
	return "cloud"
}

// CloudReadmeFile returns the README.md location for cloud integrations
func CloudReadmeFile() string {
	return path.Join(CloudDir(), "README.md")
}

// OpenstackDir returns the location for the openstack cloud integrations
func OpenstackDir() string {
	return path.Join(CloudDir(), "openstack")
}

// OpenstackReadmeFile returns the location for the openstack cloud integrations README.md
func OpenstackReadmeFile() string {
	return path.Join(OpenstackDir(), "README.md")
}

// OpenstackCloudConfFile returns the default location of the openstack cloud integrations .conf file
func OpenstackCloudConfFile() string {
	return path.Join(OpenstackDir(), "openstack.conf")
}

// OpenstackCloudConfTemplateFile returns the default location of the openstack cloud integrations .conf.template file
func OpenstackCloudConfTemplateFile() string {
	return path.Join(OpenstackDir(), "openstack.conf.template")
}

// OpenstackConfigRuntimeFile returns the location the openstack.conf is stored on nodes in the cluster
func OpenstackConfigRuntimeFile() string {
	return path.Join(constants.KubernetesDir, "openstack.conf")
}

// VSphereDir returns the location for the vsphere cloud integrations
func VSphereDir() string {
	return path.Join(CloudDir(), "vsphere")
}

// VSphereReadmeFile returns the location for the vsphere cloud integrations README.md
func VSphereReadmeFile() string {
	return path.Join(VSphereDir(), "README.md")
}

// VSphereCloudConfFile returns the default location of the vsphere cloud integrations .conf file
func VSphereCloudConfFile() string {
	return path.Join(VSphereDir(), "vsphere.conf")
}

// VSphereCloudConfTemplateFile returns the default location of the vsphere cloud integrations .conf.template file
func VSphereCloudConfTemplateFile() string {
	return path.Join(VSphereDir(), "vsphere.conf.template")
}

// VSphereConfigRuntimeFile returns the location the vsphere.conf is stored on nodes in the cluster
func VSphereConfigRuntimeFile() string {
	return path.Join(constants.KubernetesDir, "vsphere.conf")
}

// AzureDir returns the location for the azure cloud integrations
func AzureDir() string {
	return path.Join(CloudDir(), "azure")
}

// AzureReadmeFile returns the location for the azure cloud integrations README.md
func AzureReadmeFile() string {
	return path.Join(AzureDir(), "README.md")
}

// AzureCloudConfFile returns the default location of the azure cloud integrations .conf file
func AzureCloudConfFile() string {
	return path.Join(AzureDir(), "azure.conf")
}

// AzureCloudConfTemplateFile returns the default location of the azure cloud integrations .conf.template file
func AzureCloudConfTemplateFile() string {
	return path.Join(AzureDir(), "azure.conf.template")
}

// AzureConfigRuntimeFile returns the location the azure.conf is stored on nodes in the cluster
func AzureConfigRuntimeFile() string {
	return path.Join(constants.KubernetesDir, "azure.conf")
}

// AWSDir returns the location for the AWS cloud integrations
func AWSDir() string {
	return path.Join(CloudDir(), "aws")
}

// AWSReadmeFile returns the location for the AWS cloud integrations README.md
func AWSReadmeFile() string {
	return path.Join(AWSDir(), "README.md")
}

//ImageRepository returns the image registry of the cluster version
func ImageRepository(clusterVersion *version.Version) string {
	result, _ := clusterVersion.Compare("1.18.0")
	if result < 0 {
		return imageRepositoryV4
	}

	return imageRepository
}
