package bootstrap

import (
	"fmt"
	"log"
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/definitions"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

var (
	secrets = []string{
		"pki/ca.crt",
		"pki/ca.key",
		"pki/sa.key",
		"pki/sa.key",
		"pki/front-proxy-ca.crt",
		"pki/front-proxy-ca.key",
	  "pki/etcd/ca.crt",
		"pki/etcd/ca.key",
		"admin.conf",
	}
)

func Bootstrap(target string) {
	err := salt.Apply(target, &salt.Pillar{
		Bootstrap: &salt.Bootstrap{
			salt.Kubeadm{
				ConfigPath: fmt.Sprintf("salt://samples/%s/kubeadm-init.conf", definitions.CurrentDefinition()),
			},
			salt.Cni{
				ConfigPath: fmt.Sprintf("salt://samples/%s/addons/cni/flannel.yaml", definitions.CurrentDefinition()),
			},
		},
	},
		"kubelet.enable",
		"kubeadm.init",
		"cni.deploy")

	if err != nil {
		log.Fatal(err)
	}

	downloadSecrets(target)
}

func downloadSecrets(target string) {
	for _, secretLocation := range secrets {
		_, err := salt.DownloadFile(target, path.Join("/etc/kubernetes", secretLocation))
		if err != nil {
			log.Fatal(err)
		}

	}
}
