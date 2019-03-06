package etcd

import (
	"fmt"
	"log"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"

	"suse.com/caaspctl/internal/pkg/caaspctl/kubernetes"
)

func RemoveMember(node *v1.Node) error {
	masterNodes, err := kubernetes.GetMasterNodes()
	if err != nil {
		log.Fatalf("could not get the list of master nodes, aborting\n")
		return err
	}

	// Remove etcd member if target is a master
	log.Println("removing etcd member from the etcd cluster")
	for _, masterNode := range masterNodes.Items {
		log.Printf("trying to remove etcd member from master node %s\n", masterNode.ObjectMeta.Name)
		if err := RemoveMemberFrom(node, &masterNode); err == nil {
			log.Printf("etcd member for node %s removed from master node %s\n", node.ObjectMeta.Name, masterNode.ObjectMeta.Name)
			break
		} else {
			log.Printf("could not remove etcd member from master node %s\n", masterNode.ObjectMeta.Name)
		}
	}

	return nil
}

func RemoveMemberFrom(node, executorNode *v1.Node) error {
	return kubernetes.CreateAndWaitForJob(
		removeMemberFromJobName(node, executorNode),
		removeMemberFromJobSpec(node, executorNode),
	)
}

func removeMemberFromJobName(node, executorNode *v1.Node) string {
	return fmt.Sprintf("caasp-remove-etcd-member-%s-from-%s", node.ObjectMeta.Name, executorNode.ObjectMeta.Name)
}

func removeMemberFromJobSpec(node, executorNode *v1.Node) batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: removeMemberFromJobName(node, executorNode),
						// FIXME: fetch etcd image repo and tag from the clusterconfiguration in kubeadm-config configmap
						// FIXME: check that etcd member is part of the member list already
						Image: "k8s.gcr.io/etcd:3.3.10",
						Command: []string{
							"/bin/sh", "-c",
							fmt.Sprintf("etcdctl --endpoints=https://[127.0.0.1]:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key member remove $(etcdctl --endpoints=https://[127.0.0.1]:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key member list | grep ', %s,' | cut -d',' -f1)", node.ObjectMeta.Name),
						},
						Env: []v1.EnvVar{
							v1.EnvVar{
								Name:  "ETCDCTL_API",
								Value: "3",
							},
						},
						VolumeMounts: []v1.VolumeMount{
							kubernetes.VolumeMount("etc-kubernetes-pki-etcd", "/etc/kubernetes/pki/etcd", kubernetes.VolumeMountReadOnly),
						},
					},
				},
				HostNetwork:   true,
				RestartPolicy: v1.RestartPolicyNever,
				Volumes: []v1.Volume{
					kubernetes.HostMount("etc-kubernetes-pki-etcd", "/etc/kubernetes/pki/etcd"),
				},
				NodeSelector: map[string]string{
					"kubernetes.io/hostname": executorNode.ObjectMeta.Name,
				},
				Tolerations: []v1.Toleration{
					{
						Operator: v1.TolerationOpExists,
					},
				},
			},
		},
	}
}
