package deletenode

import (
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"suse.com/caaspctl/internal/pkg/caaspctl/etcd"
	"suse.com/caaspctl/internal/pkg/caaspctl/kubeadm"
	"suse.com/caaspctl/internal/pkg/caaspctl/kubernetes"
)

func DeleteNode(target string) {
	client := kubernetes.GetAdminClientSet()

	node, err := client.CoreV1().Nodes().Get(target, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("could not get node %s: %v\n", target, err)
	}

	targetName := node.ObjectMeta.Name

	var isMaster bool
	if isMaster = kubernetes.IsMaster(node); isMaster {
		log.Printf("removing master node %s\n", targetName)
	} else {
		log.Printf("removing worker node %s\n", targetName)
	}

	kubernetes.DrainNode(node)

	if isMaster {
		etcd.RemoveMember(node)
	}

	if err := kubernetes.DisarmKubelet(node); err != nil {
		log.Printf("error disarming kubelet: %v; node could be down, continuing with node removal...", err)
	}

	if isMaster {
		if err := kubeadm.RemoveAPIEndpointFromConfigMap(node); err != nil {
			log.Printf("could not remove the APIEndpoint for %s from the kubeadm-config configmap", targetName)
		}
	}

	if err := client.CoreV1().Nodes().Delete(targetName, &metav1.DeleteOptions{}); err == nil {
		log.Printf("node %s deleted successfully from the cluster\n", targetName)
	} else {
		log.Printf("could not remove node %s\n", targetName)
	}
}
