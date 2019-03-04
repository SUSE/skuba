package caaspctl

import (
	"fmt"
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

func KubeadmInitConfFile() string {
	return "kubeadm-init.conf"
}

func JoinConfDir() string {
	return "kubeadm-join.conf.d"
}

func MasterConfTemplateFile() string {
	return path.Join(JoinConfDir(), "master.conf.template")
}

func WorkerConfTemplateFile() string {
	return path.Join(JoinConfDir(), "worker.conf.template")
}

func MachineConfFile(target string) string {
	return path.Join(JoinConfDir(), fmt.Sprintf("%s.conf", target))
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
	return path.Join(AddonsDir(), "cni")
}

func FlannelManifestFile() string {
	return path.Join(CniDir(), "flannel.yaml")
}

func KubeConfigAdminFile() string {
	return "admin.conf"
}

func PkiDir() string {
	return "pki"
}
