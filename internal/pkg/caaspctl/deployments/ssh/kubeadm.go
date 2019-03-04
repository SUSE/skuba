package ssh

import (
	"github.com/pkg/errors"

	"suse.com/caaspctl/pkg/caaspctl"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

func init() {
	stateMap["kubeadm.init"] = kubeadmInit()
	stateMap["kubeadm.join"] = kubeadmJoin()
}

func kubeadmInit() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target, data interface{}) error {
		t.UploadFile(caaspctl.KubeadmInitConfFile(), "/tmp/kubeadm.conf")
		if target := sshTarget(t); target != nil {
			target.ssh("systemctl", "enable", "--now", "docker")
			target.ssh("systemctl", "stop", "kubelet")
			target.ssh("kubeadm", "init", "--config", "/tmp/kubeadm.conf", "--skip-token-print")
			target.ssh("rm", "/tmp/kubeadm.conf")
		}
		return nil
	}
	return runner
}

func kubeadmJoin() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target, data interface{}) error {
		joinConfiguration, ok := data.(deployments.JoinConfiguration)
		if !ok {
			return errors.New("couldn't access join configuration")
		}
		t.UploadFile(configPath(joinConfiguration.Role, t.Target()), "/tmp/kubeadm.conf")
		if target := sshTarget(t); target != nil {
			target.ssh("systemctl", "enable", "--now", "docker")
			target.ssh("systemctl", "stop", "kubelet")
			target.ssh("kubeadm", "join", "--config", "/tmp/kubeadm.conf")
			target.ssh("rm", "/tmp/kubeadm.conf")
		}
		return nil
	}
	return runner
}
