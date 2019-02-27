package deletenode

import (
	"fmt"
	"log"
	"os/exec"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	configutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"suse.com/caaspctl/internal/pkg/caaspctl/kubernetes"
	"suse.com/caaspctl/pkg/caaspctl"
)

func DeleteNode(target string) {
	client := kubernetes.GetAdminClientSet()

	node, err := client.CoreV1().Nodes().Get(target, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("could not get node %s: %v\n", target)
	}

	targetName := node.ObjectMeta.Name

	var isMaster bool
	if isMaster = kubernetes.IsMaster(node); isMaster {
		log.Printf("removing master node %s\n", targetName)
	} else {
		log.Printf("removing worker node %s\n", targetName)
	}

	// Drain node (shelling out, FIXME after https://github.com/kubernetes/kubernetes/pull/72827 can be used [1.14])
	cmd := exec.Command("kubectl",
		fmt.Sprintf("--kubeconfig=%s", caaspctl.KubeConfigAdminFile()),
		"drain", "--delete-local-data=true", "--force=true", "--ignore-daemonsets=true", targetName)

	if err := cmd.Run(); err != nil {
		log.Fatalf("could not drain node %s\n, aborting (use --force if you want to ignore this error)\n", targetName)
		return
	} else {
		log.Printf("node %s correctly drained\n", targetName)
	}

	if err := kubernetes.DisarmKubelet(node); err != nil {
		log.Printf("error disarming kubelet: %v; node could be down, continuing with node removal...", err)
	}

	// Perform certain actions if this node was a master
	if isMaster {
		masterNodes, err := kubernetes.GetMasterNodes()
		if err != nil {
			log.Fatalf("could not get the list of master nodes, aborting\n")
			return
		}

		// Remove master from kubeadm-config configmap ClusterStatus apiEndpoints
		kubeadmConfig, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(kubeadmconstants.KubeadmConfigConfigMap, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("could not retrieve the kubeadm-config configmap to change the apiEndpoints\n")
		}
		clusterStatus := &kubeadmapi.ClusterStatus{}
		if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), []byte(kubeadmConfig.Data[kubeadmconstants.ClusterStatusConfigMapKey]), clusterStatus); err != nil {
			log.Fatalf("could not unmarshal cluster status from kubeadm-config configmap\n")
			return
		}
		delete(clusterStatus.APIEndpoints, targetName)
		clusterStatusYaml, err := configutil.MarshalKubeadmConfigObject(clusterStatus)
		if err != nil {
			log.Fatalf("could not marshal modified cluster status\n")
			return
		}
		_, err = client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Update(&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kubeadmconstants.KubeadmConfigConfigMap,
				Namespace: metav1.NamespaceSystem,
			},
			Data: map[string]string{
				kubeadmconstants.ClusterConfigurationConfigMapKey: kubeadmConfig.Data[kubeadmconstants.ClusterConfigurationConfigMapKey],
				kubeadmconstants.ClusterStatusConfigMapKey:        string(clusterStatusYaml),
			},
		})
		if err != nil {
			log.Fatalf("could not update kubeadm-config configmap\n")
			return
		}

		// Remove etcd member if target is a master
		log.Println("removing etcd member from the etcd cluster")
		for _, masterNode := range masterNodes.Items {
			log.Printf("trying to remove etcd member from master node %s\n", masterNode.ObjectMeta.Name)
			if err := kubernetes.RemoveEtcdMember(node, &masterNode); err == nil {
				log.Printf("etcd member removed from master node %s\n", masterNode.ObjectMeta.Name)
				break
			} else {
				log.Printf("could not remove etcd member from master node %s\n", masterNode.ObjectMeta.Name)
			}
		}
	}

	// Delete node
	if err := client.CoreV1().Nodes().Delete(targetName, &metav1.DeleteOptions{}); err == nil {
		log.Printf("node %s deleted successfully from the cluster\n", targetName)
	} else {
		log.Printf("could not remove node %s\n", targetName)
	}
}
