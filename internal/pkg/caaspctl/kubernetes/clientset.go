package kubernetes

import (
	"log"

	clientset "k8s.io/client-go/kubernetes"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"suse.com/caaspctl/pkg/caaspctl"
)

func GetAdminClientSet() *clientset.Clientset {
	client, err := kubeconfigutil.ClientSetFromFile(caaspctl.KubeConfigAdminFile())
	if err != nil {
		log.Fatal("could not load admin kubeconfig file")
	}
	return client
}
