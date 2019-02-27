package kubernetes

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	masterLabel = "node-role.kubernetes.io/master"
)

func GetMasterNodes() (*v1.NodeList, error) {
	return GetAdminClientSet().CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=", masterLabel),
	})
}

func IsMaster(node *v1.Node) bool {
	_, isMaster := node.ObjectMeta.Labels[masterLabel]
	return isMaster
}
