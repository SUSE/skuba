package bootstrap

import (
	"io/ioutil"
	"log"
	"os"
	"path"

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

func Bootstrap(target salt.Target, masterConfig salt.MasterConfig) {
	err := salt.Apply(target, masterConfig, &salt.Pillar{
		Bootstrap: &salt.Bootstrap{
			salt.Kubeadm{
				ConfigPath: "salt://kubeadm-init.conf",
			},
			salt.Cni{
				ConfigPath: "salt://addons/cni/flannel.yaml",
			},
		},
	},
		"kubelet.enable",
		"kubeadm.init",
		"cni.deploy")

	if err != nil {
		log.Fatal(err)
	}

	downloadSecrets(target, masterConfig)
}

func downloadSecrets(target salt.Target, masterConfig salt.MasterConfig) {
	os.MkdirAll(path.Join("pki", "etcd"), 0755)

	for _, secretLocation := range secrets {
		secretData, err := salt.DownloadFile(
			target,
			masterConfig,
			path.Join("/etc/kubernetes", secretLocation))
		if err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(path.Join("pki", secretLocation), []byte(secretData), 0644); err != nil {
			log.Fatal(err)
		}
	}
}
