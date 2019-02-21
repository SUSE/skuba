package bootstrap

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/definitions"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

var (
	secrets = []string{
		"pki/ca.crt",
		"pki/ca.key",
		"pki/sa.key",
		"pki/sa.pub",
		"pki/front-proxy-ca.crt",
		"pki/front-proxy-ca.key",
		"pki/etcd/ca.crt",
		"pki/etcd/ca.key",
		"admin.conf",
	}
)

func Bootstrap(target salt.Target) {
	err := salt.Apply(target, &salt.Pillar{
		Bootstrap: &salt.Bootstrap{
			salt.Kubeadm{
				ConfigPath: fmt.Sprintf("%s/%s", salt.CurrentDefinitionPrefix(), "kubeadm-init.conf"),
			},
			salt.Cni{
				ConfigPath: fmt.Sprintf("%s/%s", salt.CurrentDefinitionPrefix(), "addons/cni/flannel.yaml"),
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

func downloadSecrets(target salt.Target) {
	secretPrefix := definitions.PKIPath()
	os.MkdirAll(path.Join(secretPrefix, "pki", "etcd"), 0755)

	for _, secretLocation := range secrets {
		secretData, err := salt.DownloadFile(
			target,
			path.Join("/etc/kubernetes", secretLocation))
		if err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(path.Join(secretPrefix, secretLocation), []byte(secretData), 0644); err != nil {
			log.Fatal(err)
		}
	}
}
