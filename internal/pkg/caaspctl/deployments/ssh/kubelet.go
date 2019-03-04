package ssh

import (
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/ssh/assets"
)

func init() {
	stateMap["kubelet.configure"] = kubeletConfigure()
	stateMap["kubelet.enable"] = kubeletEnable()
}

func kubeletConfigure() Runner {
	runner := struct{ State }{}
	runner.DoRun = func(t *Target, data interface{}) error {
		osRelease, _ := t.OSRelease()
		if osRelease["ID_LIKE"] == "debian" {
			t.UploadFileContents("/lib/systemd/system/kubelet.service", assets.KubeletService)
			t.UploadFileContents("/etc/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService)
		} else {
			t.UploadFileContents("/usr/lib/systemd/system/kubelet.service", assets.KubeletService)
			t.UploadFileContents("/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService)
			t.UploadFileContents("/etc/sysconfig/kubelet", assets.KubeletSysconfig)
		}
		t.ssh("systemctl", "daemon-reload")
		return nil
	}
	return runner
}

func kubeletEnable() Runner {
	runner := struct{ State }{}
	runner.DoRun = func(t *Target, data interface{}) error {
		t.ssh("systemctl", "enable", "kubelet")
		return nil
	}
	return runner
}
