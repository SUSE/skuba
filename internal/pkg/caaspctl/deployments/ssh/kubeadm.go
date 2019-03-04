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

func kubeadmInit() Runner {
	runner := struct{ State }{}
	runner.DoRun = func(t *Target, data interface{}) error {
		t.Target.UploadFile(caaspctl.KubeadmInitConfFile(), "/tmp/kubeadm.conf")
		t.ssh("systemctl", "enable", "--now", "docker")
		t.ssh("systemctl", "stop", "kubelet")
		t.ssh("kubeadm", "init", "--config", "/tmp/kubeadm.conf", "--skip-token-print")
		t.ssh("rm", "/tmp/kubeadm.conf")
		return nil
	}
	return runner
}

func kubeadmJoin() Runner {
	runner := struct{ State }{}
	runner.DoRun = func(t *Target, data interface{}) error {
		joinConfiguration, ok := data.(deployments.JoinConfiguration)
		if !ok {
			return errors.New("couldn't access join configuration")
		}
		t.Target.UploadFile(configPath(joinConfiguration.Role, t.Target.Target()), "/tmp/kubeadm.conf")
		t.ssh("systemctl", "enable", "--now", "docker")
		t.ssh("systemctl", "stop", "kubelet")
		t.ssh("kubeadm", "join", "--config", "/tmp/kubeadm.conf")
		t.ssh("rm", "/tmp/kubeadm.conf")
		return nil
	}
	return runner
}
