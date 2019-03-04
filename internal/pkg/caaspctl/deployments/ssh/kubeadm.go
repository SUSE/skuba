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
		if err := t.UploadFile(caaspctl.KubeadmInitConfFile(), "/tmp/kubeadm.conf"); err != nil {
			return err
		}
		if _, _, err := t.ssh("systemctl", "enable", "--now", "docker"); err != nil {
			return err
		}
		if _, _, err := t.ssh("systemctl", "stop", "kubelet"); err != nil {
			return err
		}
		if _, _, err := t.ssh("kubeadm", "init", "--config", "/tmp/kubeadm.conf", "--skip-token-print"); err != nil {
			return err
		}
		if _, _, err := t.ssh("rm", "/tmp/kubeadm.conf"); err != nil {
			return err
		}
		return nil
	}
}

func kubeadmJoin() Runner {
	return func(t *Target, data interface{}) error {
		joinConfiguration, ok := data.(deployments.JoinConfiguration)
		if !ok {
			return errors.New("couldn't access join configuration")
		}
		if err := t.UploadFile(configPath(joinConfiguration.Role, t.Node()), "/tmp/kubeadm.conf"); err != nil {
			return err
		}
		if _, _, err := t.ssh("systemctl", "enable", "--now", "docker"); err != nil {
			return err
		}
		if _, _, err := t.ssh("systemctl", "stop", "kubelet"); err != nil {
			return err
		}
		if _, _, err := t.ssh("kubeadm", "join", "--config", "/tmp/kubeadm.conf"); err != nil {
			return err
		}
		if _, _, err := t.ssh("rm", "/tmp/kubeadm.conf"); err != nil {
			return err
		}
		return nil
	}
}
