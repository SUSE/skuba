/*
 * Copyright (c) 2019 SUSE LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package etcd

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"

	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/pkg/errors"
)

func RemoveMember(node *v1.Node) error {
	masterNodes, err := kubernetes.GetMasterNodes()
	if err != nil {
		return errors.Wrap(err, "could not get the list of master nodes, aborting")
	}

	// Remove etcd member if target is a master
	klog.V(1).Info("removing etcd member from the etcd cluster")
	for _, masterNode := range masterNodes.Items {
		klog.V(1).Infof("trying to remove etcd member from master node %s", masterNode.ObjectMeta.Name)
		if err := RemoveMemberFrom(node, &masterNode); err == nil {
			klog.V(1).Infof("etcd member for node %s removed from master node %s", node.ObjectMeta.Name, masterNode.ObjectMeta.Name)
			break
		} else {
			klog.V(1).Infof("could not remove etcd member from master node %s", masterNode.ObjectMeta.Name)
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
							{
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
