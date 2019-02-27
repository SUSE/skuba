package etcd

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"

	"suse.com/caaspctl/internal/pkg/caaspctl/kubernetes"
)

func RemoveMember(node, executorNode *v1.Node) error {
	return kubernetes.CreateAndWaitForJob(removeMember(node, executorNode))
}

func removeMember(node, executorNode *v1.Node) (string, batchv1.JobSpec) {
	return removeMemberJobName(node, executorNode), removeMemberJobSpec(node, executorNode)
}

func removeMemberJobName(node, executorNode *v1.Node) string {
	return fmt.Sprintf("caasp-remove-etcd-member-%s-from-%s", node.ObjectMeta.Name, executorNode.ObjectMeta.Name)
}

func removeMemberJobSpec(node, executorNode *v1.Node) batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: removeMemberJobName(node, executorNode),
						// FIXME: fetch etcd image repo and tag from the clusterconfiguration in kubeadm-config configmap
						// FIXME: check that etcd member is part of the member list already
						Image: "k8s.gcr.io/etcd:3.3.10",
						Command: []string{
							"/bin/sh", "-c",
							fmt.Sprintf("etcdctl --endpoints=https://[127.0.0.1]:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key member remove $(etcdctl --endpoints=https://[127.0.0.1]:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key member list | grep ', %s' | cut -d',' -f1)", node.ObjectMeta.Name),
						},
						Env: []v1.EnvVar{
							v1.EnvVar{
								Name:  "ETCDCTL_API",
								Value: "3",
							},
						},
						VolumeMounts: []v1.VolumeMount{
							kubernetes.VolumeMount("etc-kubernetes-pki", "/etc/kubernetes/pki"),
						},
					},
				},
				HostNetwork:   true,
				RestartPolicy: v1.RestartPolicyNever,
				Volumes: []v1.Volume{
					kubernetes.HostMount("etc-kubernetes-pki", "/etc/kubernetes/pki"),
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
