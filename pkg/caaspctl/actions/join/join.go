package join

import (
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
	"suse.com/caaspctl/pkg/caaspctl"
)

type JoinConfiguration struct {
	Role caaspctl.Role
}

func Join(joinConfiguration JoinConfiguration, target deployments.Target) {
	statesToApply := []string{"kubelet.configure", "kubelet.enable", "kubeadm.join"}

	if joinConfiguration.Role == caaspctl.MasterRole {
		statesToApply = append([]string{"kubernetes.upload-secrets"}, statesToApply...)
	}

	target.Apply(statesToApply...)
}
