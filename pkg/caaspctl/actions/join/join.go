package join

import (
	"fmt"
	"log"
	"os"
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

func Join(target string) {
	salt.Apply(target, &salt.Pillar{
		Kubeadm: &salt.Kubeadm{
			ConfigPath: configPath(target),
		},
	},
		"kubelet.enable",
		"kubeadm.join")
}

func configPath(target string) string {
	targetPath := path.Join("samples/3-masters-3-workers-vagrant/kubeadm-join-conf.d", fmt.Sprintf("%s.conf", target))
	if _, err := os.Stat(path.Join("deployments/salt/states", targetPath)); err == nil {
		return fmt.Sprintf("salt://%s", targetPath)
	} else {
		log.Fatal(err)
	}
	return ""
}
