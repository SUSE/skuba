package kubernetes

import (
	v1 "k8s.io/api/core/v1"
)

func RemoveEtcdMember(node, executorNode *v1.Node) error {
	return nil
}
