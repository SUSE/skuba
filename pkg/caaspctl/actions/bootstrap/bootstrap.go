package bootstrap

import (
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

func Bootstrap(target string) {
	salt.Apply(target, &salt.Pillar{
		&salt.Kubeadm{
			ConfigPath: "salt://samples/3-masters-3-workers-vagrant/kubeadm-init.conf",
		},
		&salt.Cni{
			ConfigPath: "salt://samples/3-masters-3-workers-vagrant/addons/cni/flannel.yaml",
		},
	},
	"kubelet.enable",
	"kubeadm.init",
	"cni.deploy")
}
