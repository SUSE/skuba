package kubernetes

import (
	"fmt"
	"log"
	"os/exec"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"suse.com/caaspctl/pkg/caaspctl"
)

func GetMasterNodes() (*v1.NodeList, error) {
	return GetAdminClientSet().CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=", kubeadmconstants.LabelNodeRoleMaster),
	})
}

func IsMaster(node *v1.Node) bool {
	_, isMaster := node.ObjectMeta.Labels[kubeadmconstants.LabelNodeRoleMaster]
	return isMaster
}

func DrainNode(node *v1.Node) error {
	// Drain node (shelling out, FIXME after https://github.com/kubernetes/kubernetes/pull/72827 can be used [1.14])
	cmd := exec.Command("kubectl",
		fmt.Sprintf("--kubeconfig=%s", caaspctl.KubeConfigAdminFile()),
		"drain", "--delete-local-data=true", "--force=true", "--ignore-daemonsets=true", node.ObjectMeta.Name)

	if err := cmd.Run(); err != nil {
		log.Printf("could not drain node %s, aborting (use --force if you want to ignore this error)\n", node.ObjectMeta.Name)
		return err
	} else {
		log.Printf("node %s correctly drained\n", node.ObjectMeta.Name)
	}

	return nil
}
