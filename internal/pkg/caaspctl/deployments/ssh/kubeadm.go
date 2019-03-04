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
	return func(t *Target, data interface{}) error {
		t.UploadFile(caaspctl.KubeadmInitConfFile(), "/tmp/kubeadm.conf")
		t.ssh("systemctl", "enable", "--now", "docker")
		t.ssh("systemctl", "stop", "kubelet")
		t.ssh("kubeadm", "init", "--config", "/tmp/kubeadm.conf", "--skip-token-print")
		t.ssh("rm", "/tmp/kubeadm.conf")
		return nil
	}
}

func kubeadmJoin() Runner {
	return func(t *Target, data interface{}) error {
		joinConfiguration, ok := data.(deployments.JoinConfiguration)
		if !ok {
			return errors.New("couldn't access join configuration")
		}
		t.UploadFile(configPath(joinConfiguration.Role, t.Node()), "/tmp/kubeadm.conf")
		t.ssh("systemctl", "enable", "--now", "docker")
		t.ssh("systemctl", "stop", "kubelet")
		t.ssh("kubeadm", "join", "--config", "/tmp/kubeadm.conf")
		t.ssh("rm", "/tmp/kubeadm.conf")
		return nil
	}
}
