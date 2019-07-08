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
	"path/filepath"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
)

const (
	CRISocket  = "/var/run/crio/crio.sock"
	SUSECNIDir = "/usr/lib/cni"
)

func KubeadmInitConfFile() string {
	return "kubeadm-init.conf"
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

func CniDir() string {
	return filepath.Join(AddonsDir(), "cni")
}

func CiliumManifestFile() string {
	return filepath.Join(CniDir(), "cilium.yaml")
}

func PspDir() string {
	return filepath.Join(AddonsDir(), "psp")
}

func PspUnprivManifestFile() string {
	return filepath.Join(PspDir(), "podsecuritypolicy-unprivileged.yaml")
}

func PspPrivManifestFile() string {
	return filepath.Join(PspDir(), "podsecuritypolicy-privileged.yaml")
}

func KubeConfigAdminFile() string {
	return "admin.conf"
}

func PkiDir() string {
	return "pki"
}

func KuredDir() string {
	return filepath.Join(AddonsDir(), "kured")
}

func KuredManifestFile() string {
	return filepath.Join(KuredDir(), "kured.yaml")
}

func DexDir() string {
	return filepath.Join(AddonsDir(), "dex")
}

func DexManifestFile() string {
	return filepath.Join(DexDir(), "dex.yaml")
}

func GangwayDir() string {
	return filepath.Join(AddonsDir(), "gangway")
}

func GangwayManifestfile() string {
	return filepath.Join(GangwayDir(), "gangway.yaml")
}
