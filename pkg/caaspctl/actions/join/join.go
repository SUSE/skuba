package join

import (
	"fmt"
	"log"
	"os"
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

const (
	MasterRole = iota
	WorkerRole = iota
)

type Role int

type JoinConfiguration struct {
	Target salt.Target
	Role   Role
}

func Join(joinConfiguration JoinConfiguration, masterConfig salt.MasterConfig) {
	statesToApply := []string{"kubelet.enable", "kubeadm.join"}

	pillar := &salt.Pillar{
		Join: &salt.Join{
			Kubeadm: salt.Kubeadm{
				ConfigPath: configPath(joinConfiguration.Target.Node),
			},
		},
	}

	if joinConfiguration.Role == MasterRole {
		statesToApply = append([]string{"kubernetes.upload-secrets"}, statesToApply...)
		pillar.Join.Kubernetes = &salt.Kubernetes{
			AdminConfPath: "salt://admin.conf",
			SecretsPath:   "salt://pki",
		}
	}

	salt.Apply(joinConfiguration.Target, masterConfig, pillar, statesToApply...)
}

func configPath(target string) string {
	targetPath := path.Join(
		"kubeadm-join-conf.d",
		fmt.Sprintf("%s.conf", target),
	)
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Sprintf("salt://%s", targetPath)
	} else {
		log.Fatal(err)
	}
	return ""
}
