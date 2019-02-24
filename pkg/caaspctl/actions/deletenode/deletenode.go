package deletenode

import (
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
	"suse.com/caaspctl/internal/pkg/caaspctl/kubernetes"
)

func DeleteNode(masterConfig salt.MasterConfig) {
	if err := salt.Apply(masterConfig, nil, "kubelet.disable"); err != nil {
		log.Println("could not disable the kubelet, continuing with node removal...")
	}
	client := kubernetes.GetAdminClientSet()
	node, err := client.CoreV1().Nodes().Get(masterConfig.Target.Node, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("could not get node %s: %v", masterConfig.Target.Node, err)
	}
	if _, isMaster := node.ObjectMeta.Labels["node-role.kubernetes.io/master"]; isMaster {
		log.Printf("removing master node %s\n", masterConfig.Target.Node)
		// TODO: remove etcd member as well; target node might be down though
	} else {
		log.Printf("removing worker node %s\n", masterConfig.Target.Node)
	}
	if err := client.CoreV1().Nodes().Delete(masterConfig.Target.Node, &metav1.DeleteOptions{}); err == nil {
		log.Printf("node %s deleted successfully from the cluster\n", masterConfig.Target.Node)
	} else {
		log.Printf("could not remove node %s\n", masterConfig.Target.Node)
	}
}
