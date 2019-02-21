package bootstrap

import (
	"fmt"
	"log"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

func Bootstrap(target string) {
	err := salt.Apply(target, &salt.Pillar{
		Bootstrap: &salt.Bootstrap{
			salt.Kubeadm{
				ConfigPath: "salt://samples/3-masters-3-workers-vagrant/kubeadm-init.conf",
			},
			salt.Cni{
				ConfigPath: "salt://samples/3-masters-3-workers-vagrant/addons/cni/flannel.yaml",
			},
		},
	},
		"kubelet.enable",
		"kubeadm.init",
		"cni.deploy")

	if err != nil {
		log.Fatal(err)
	}

	caCrt, _ := salt.DownloadFile(target, "/etc/kubernetes/pki/ca.crt")
	caKey, _ := salt.DownloadFile(target, "/etc/kubernetes/pki/ca.key")
	saKey, _ := salt.DownloadFile(target, "/etc/kubernetes/pki/sa.key")
	saPubKey, _ := salt.DownloadFile(target, "/etc/kubernetes/pki/sa.pub")
	frontProxyCaCrt, _ := salt.DownloadFile(target, "/etc/kubernetes/pki/front-proxy-ca.crt")
	frontProxyCaKey, _ := salt.DownloadFile(target, "/etc/kubernetes/pki/front-proxy-ca.key")
	etcdCaCrt, _ := salt.DownloadFile(target, "/etc/kubernetes/pki/etcd/ca.crt")
	etcdCaKey, _ := salt.DownloadFile(target, "/etc/kubernetes/pki/etcd/ca.key")
	adminConf, _ := salt.DownloadFile(target, "/etc/kubernetes/admin.conf")

	secrets := []string{caCrt, caKey, saKey, saPubKey, frontProxyCaCrt,
		frontProxyCaKey, etcdCaCrt, etcdCaKey, adminConf}

	for _, secret := range secrets {
		fmt.Printf("===\nSecret retrieved:\n%s\n", secret)
	}
}
