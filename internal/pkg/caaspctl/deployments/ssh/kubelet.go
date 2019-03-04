package ssh

import (
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/ssh/assets"
)

func init() {
	stateMap["kubelet.configure"] = kubeletConfigure()
	stateMap["kubelet.enable"] = kubeletEnable()
}

func kubeletConfigure() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target, data interface{}) error {
		t.UploadFileContents("/lib/systemd/system/kubelet.service", assets.KubeletService)
		t.UploadFileContents("/etc/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService)
		t.UploadFileContents("/etc/sysconfig/kubelet", assets.KubeletSysconfig)
		if target := sshTarget(t); target != nil {
			target.ssh("systemctl", "daemon-reload")
		}
		return nil
	}
	return runner
}

func kubeletEnable() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target, data interface{}) error {
		if target := sshTarget(t); target != nil {
			target.ssh("systemctl", "enable", "kubelet")
		}
		return nil
	}
	return runner
}
