package ssh

import (
	"github.com/pkg/errors"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
	"suse.com/caaspctl/pkg/caaspctl"
	node "suse.com/caaspctl/pkg/caaspctl/actions/node/join"
)

func init() {
	stateMap["kubeadm.init"] = kubeadmInit()
	stateMap["kubeadm.join"] = kubeadmJoin()
}

func kubeadmInit() Runner {
	return func(t *Target, data interface{}) error {
		defer t.ssh("rm", "/tmp/kubeadm.conf")

		if err := t.target.UploadFile(caaspctl.KubeadmInitConfFile(), "/tmp/kubeadm.conf"); err != nil {
			return err
		}
		if _, _, err := t.ssh("systemctl", "enable", "--now", "docker"); err != nil {
			return err
		}
		if _, _, err := t.ssh("systemctl", "stop", "kubelet"); err != nil {
			return err
		}
		_, _, err := t.ssh("kubeadm", "init", "--config", "/tmp/kubeadm.conf", "--skip-token-print")
		return err
	}
}

func kubeadmJoin() Runner {
	return func(t *Target, data interface{}) error {
		joinConfiguration, ok := data.(deployments.JoinConfiguration)
		if !ok {
			return errors.New("couldn't access join configuration")
		}

		defer t.ssh("rm", "/tmp/kubeadm.conf")

		if err := t.target.UploadFile(node.ConfigPath(joinConfiguration.Role, t.target), "/tmp/kubeadm.conf"); err != nil {
			return err
		}
		if _, _, err := t.ssh("systemctl", "enable", "--now", "docker"); err != nil {
			return err
		}
		if _, _, err := t.ssh("systemctl", "stop", "kubelet"); err != nil {
			return err
		}
		_, _, err := t.ssh("kubeadm", "join", "--config", "/tmp/kubeadm.conf")
		return err
	}
}
