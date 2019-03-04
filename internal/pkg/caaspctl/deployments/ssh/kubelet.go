package ssh

import (
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/ssh/assets"
)

func init() {
	stateMap["kubelet.configure"] = kubeletConfigure()
	stateMap["kubelet.enable"] = kubeletEnable()
}

func kubeletConfigure() Runner {
	return func(t *Target, data interface{}) error {
		osRelease, _ := t.OSRelease()
		if osRelease["ID_LIKE"] == "debian" {
			if err := t.UploadFileContents("/lib/systemd/system/kubelet.service", assets.KubeletService); err != nil {
				return err
			}
			if err := t.UploadFileContents("/etc/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService); err != nil {
				return err
			}
		} else {
			if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service", assets.KubeletService); err != nil {
				return err
			}
			if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService); err != nil {
				return err
			}
			if err := t.UploadFileContents("/etc/sysconfig/kubelet", assets.KubeletSysconfig); err != nil {
				return err
			}
		}
		_, _, err := t.ssh("systemctl", "daemon-reload")
		return err
	}
}

func kubeletEnable() Runner {
	return func(t *Target, data interface{}) error {
		_, _, err := t.ssh("systemctl", "enable", "kubelet")
		return err
	}
}
