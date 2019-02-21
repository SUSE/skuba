package join

import (
	"fmt"
	"log"
	"os"
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/constants"
	"suse.com/caaspctl/internal/pkg/caaspctl/definitions"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

const (
	MasterRole = iota
	WorkerRole = iota
)
type Role int

func Join(target salt.Target, role Role) {
	statesToApply := []string{"kubelet.enable", "kubeadm.join"}

	pillar := &salt.Pillar{
		Join: &salt.Join{
			Kubeadm: salt.Kubeadm{
				ConfigPath: configPath(target.Node),
			},
		},
	}

	if role == MasterRole {
		statesToApply = append([]string{"kubernetes.upload-secrets"}, statesToApply...)
		pillar.Join.Kubernetes = &salt.Kubernetes{
			SecretsPath: secretsPath(),
		}
	}

	salt.Apply(target, pillar, statesToApply...)
}

func configPath(target string) string {
	targetPath := path.Join(
		definitions.CurrentDefinitionPrefix(),
		"kubeadm-join-conf.d",
		fmt.Sprintf("%s.conf", target),
	)
	if _, err := os.Stat(path.Join(constants.DefinitionPath, "states", targetPath)); err == nil {
		return fmt.Sprintf("salt://%s", targetPath)
	} else {
		log.Fatal(err)
	}
	return ""
}

func secretsPath() string {
	return fmt.Sprintf("%s/%s", salt.CurrentDefinitionPrefix(), "pki")
}
