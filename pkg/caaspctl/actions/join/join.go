package join

import (
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

func Join(joinConfiguration deployments.JoinConfiguration, target deployments.Target) {
	statesToApply := []string{"kubelet.configure", "kubelet.enable", "kubeadm.join"}

	if joinConfiguration.Role == deployments.MasterRole {
		statesToApply = append([]string{"kubernetes.upload-secrets"}, statesToApply...)
	}

	target.Apply(joinConfiguration, statesToApply...)
}
