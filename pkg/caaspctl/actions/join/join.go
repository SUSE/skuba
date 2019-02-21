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

func Join(target string) {
	salt.Apply(target, &salt.Pillar{
		Join: &salt.Join{
			Kubeadm: salt.Kubeadm{
				ConfigPath: configPath(target),
			},
		},
	},
		"kubelet.enable",
		"kubeadm.join")
}

func configPath(target string) string {
	targetPath := path.Join(fmt.Sprintf("samples/%s/kubeadm-join-conf.d", definitions.CurrentDefinition()), fmt.Sprintf("%s.conf", target))
	if _, err := os.Stat(path.Join(constants.DefinitionPath, "states", targetPath)); err == nil {
		return fmt.Sprintf("salt://%s", targetPath)
	} else {
		log.Fatal(err)
	}
	return ""
}
