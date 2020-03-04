/*
 * Copyright (c) 2020 SUSE LLC.
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

package replica

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/pkg/errors"
)

type affinity int

const (
	preferred affinity = iota
	required
)

const (
	highAvailabilitylabel = "caasp.suse.com/skuba-replica-ha"
	minSize               = 2
	retryCount            = 5
	retryInterval         = kubeadmconstants.DiscoveryRetryInterval
	retryTimeout          = kubeadmconstants.PatchNodeTimeout
	patchAffinityRemove   = `[
		{
			"op":"remove",
			"path":"/spec/template/spec/affinity"
		}
	]`
	patchAffinityRequired = `{
		"spec": {
			"template": {
				"spec": {
					"affinity": {
						"podAntiAffinity": {
							"requiredDuringSchedulingIgnoredDuringExecution": [
								{
									"labelSelector": {
										"matchExpressions": [
											{
												"key": "app",
												"operator": "In",
												"values": [
													"%s"
												]
											}
										]
									},
									"topologyKey": "kubernetes.io/hostname"
								}
							]
						}
					}
				}
			}
		}
	}`
	patchAffinityPreferred = `{
		"spec": {
			"template": {
				"spec": {
					"affinity": {
						"podAntiAffinity": {
							"preferredDuringSchedulingIgnoredDuringExecution": [
								{
									"weight": 100,
									"podAffinityTerm": {
										"labelSelector": {
											"matchExpressions": [
												{
													"key": "app",
													"operator": "In",
													"values": [
														"%s"
													]
												}
											]
										},
										"topologyKey": "kubernetes.io/hostname"
									}
								}
							]
						}
					}
				}
			}
		}
	}`
	patchReplicas = `{
		"spec": {
			"replicas": %v
		}
	}`
)

// Helper provide methods for replica update of deployment
// resource that has highAvailabilitylabel.
type Helper struct {
	client      clientset.Interface
	clusterSize int
	deployments *appsv1.DeploymentList
	deployment  *appsv1.Deployment
	replicaSize int
	minSize     int
}

// NewHelper creates a helper to update deployment replicas.
func NewHelper(client clientset.Interface) (*Helper, error) {
	node, err := kubernetes.GetAllNodes(client)
	if err != nil {
		return nil, err
	}

	return &Helper{
		client:      client,
		clusterSize: len(node.Items),
		minSize:     minSize,
	}, nil
}

// deploymentsHelper updates deployment list in Helper object
func (r *Helper) deploymentsHelper() error {
	deployments, err := r.client.AppsV1().Deployments(metav1.NamespaceSystem).List(
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=true", highAvailabilitylabel),
		},
	)
	if err != nil {
		return err
	}
	r.deployments = deployments
	return nil
}

// deploymentHelper updates deployment in Helper object
func (r *Helper) deploymentHelper(name string) error {
	deployment, err := r.client.AppsV1().Deployments(metav1.NamespaceSystem).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	r.deployment = deployment
	return nil
}

// replaceAffinity updates affinity rules and triggers pod reschedule
func (r *Helper) replaceAffinity(remove affinity, create affinity) (bool, error) {
	// doRemove will remove only when affinity path exists
	doRemove := false
	affinity := r.deployment.Spec.Template.Spec.Affinity
	if affinity != nil && affinity.PodAntiAffinity != nil {
		preferredList := len(r.deployment.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
		requiredList := len(r.deployment.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
		switch remove {
		case preferred:
			// when preferred affinity exist it will be removed first
			if preferredList != 0 {
				doRemove = true
				break
			}
			// no op when required affinity already exist.
			if requiredList != 0 {
				return false, nil
			}
		case required:
			// when required affinity exist it will be removed first
			if requiredList != 0 {
				doRemove = true
				break
			}
			// no op when preferred affinity already exist.
			if preferredList != 0 {
				return false, nil
			}
		}
	}
	if doRemove {
		if err := r.removeAffinity(); err != nil {
			return false, err
		}
	}

	var affinityJSON string
	switch create {
	case required:
		affinityJSON = fmt.Sprintf(patchAffinityRequired, r.deployment.ObjectMeta.Name)
	default:
		affinityJSON = fmt.Sprintf(patchAffinityPreferred, r.deployment.ObjectMeta.Name)
	}

	// At least one node needs to be free of the affinity rules for pods to re-distribute.
	// Sizing down replicas would make room for new pod resource.
	// For example, during rolling upgrade if all node already has `required`, scheduler will `Pend`
	// when tries to create a new deployment with `preferred`. This is due to the conflict to its original rule.
	// But this ends up with less of replicas. So need the following patch to bring the replicas to the right size
	// after affinity updated.
	if _, err := r.updateDeploymentReplica(r.clusterSize - 1); err != nil {
		return false, err
	}
	_, err := r.client.AppsV1().Deployments(metav1.NamespaceSystem).Patch(r.deployment.ObjectMeta.Name, types.StrategicMergePatchType, []byte(affinityJSON))
	if err != nil {
		return false, err
	}
	if _, err := r.updateDeploymentReplica(r.replicaSize); err != nil {
		return false, err
	}

	return true, nil
}

// removeAffinity patch to remove affinity from deployment
func (r *Helper) removeAffinity() error {
	_, err := r.client.AppsV1().Deployments(metav1.NamespaceSystem).Patch(r.deployment.ObjectMeta.Name, types.JSONPatchType, []byte(patchAffinityRemove))
	if err != nil {
		return err
	}
	return nil
}

// updateDeploymentReplic patch to update replica size of deployment
func (r *Helper) updateDeploymentReplica(size int) (*appsv1.Deployment, error) {
	replicaJSON := fmt.Sprintf(patchReplicas, size)
	return r.client.AppsV1().Deployments(metav1.NamespaceSystem).Patch(r.deployment.ObjectMeta.Name, types.StrategicMergePatchType, []byte(replicaJSON))
}

// UpdateNodes patches for replicas affinity rules after adding nodes, before removing nodes, or after addon
// upgrade. This updates affinity rule to preferredDuringSchedulingIgnoredDuringExecution of cluster size
// less then replica size, and updates affinity rule to requiredDuringSchedulingIgnoredDuringExecution when
// cluster size is equal or greater to replicas.
func (r *Helper) UpdateNodes() error {
	if err := r.deploymentsHelper(); err != nil {
		return err
	}

	for _, deployment := range r.deployments.Items {
		if err := r.deploymentHelper(deployment.Name); err != nil {
			return err
		}
		r.replicaSize = int(*r.deployment.Spec.Replicas)
		if r.replicaSize == 0 {
			r.replicaSize = r.minSize
		}

		updated := false
		var err error
		switch {
		case r.clusterSize >= r.replicaSize:
			updated, err = r.replaceAffinity(preferred, required)
			if err != nil {
				return err
			}
		case r.clusterSize >= r.minSize:
			updated, err = r.replaceAffinity(required, preferred)
			if err != nil {
				return err
			}
		}
		if !updated && r.clusterSize <= r.replicaSize {
			// update replicas to trigger pod re-distribution when affinity is in
			// preferredDuringSchedulingIgnoredDuringExecution.
			if _, err := r.updateDeploymentReplica(r.clusterSize - 1); err != nil {
				return err
			}
			if _, err := r.updateDeploymentReplica(r.replicaSize); err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateBeforeNodeDrains patches affinity rules and wait for pod running before node remove.
// Patch will apply to deployment with preferredDuringSchedulingIgnoredDuringExecution affinity
// when cluster size equal or less then replicas.
func (r *Helper) UpdateBeforeNodeDrains() error {
	if err := r.deploymentsHelper(); err != nil {
		return err
	}

	for _, deployment := range r.deployments.Items {
		if err := r.deploymentHelper(deployment.Name); err != nil {
			return err
		}
		r.replicaSize = int(*r.deployment.Spec.Replicas)
		if r.clusterSize > r.replicaSize {
			continue
		}

		updated, err := r.replaceAffinity(required, preferred)
		if err != nil {
			return err
		}
		if !updated {
			// update replicas to trigger pod re-distribution when affinity is in
			// preferredDuringSchedulingIgnoredDuringExecution.
			if _, err := r.updateDeploymentReplica(r.clusterSize - 1); err != nil {
				return err
			}
			if _, err := r.updateDeploymentReplica(r.replicaSize); err != nil {
				return err
			}
			updated = true
		}
		if updated {
			var e error
			retry := retryCount
			if err := r.waitForUpdates(retry, e); err != nil {
				return err
			}
			klog.Infof("deployment %s is available", deployment.Name)
		}
	}

	return nil
}

func (r *Helper) waitForUpdates(retry int, newErr error) error {
	switch {
	case retry == 0:
		return errors.Wrap(newErr, "retry exhausted")
	default:
		klog.Warningf("waiting for deployment be available: %d", retry)
		retry--
		time.Sleep(retryInterval)

		if newErr = r.waitForDeploymentReplicas(); newErr == nil {
			return nil
		}

		if err := r.removePendingPods(); err != nil {
			return err
		}

		if err := r.waitForUpdates(retry, newErr); err != nil {
			return err
		}
	}
	return nil
}

func (r *Helper) waitForDeploymentReplicas() error {
	return wait.PollImmediate(retryInterval, retryTimeout, func() (bool, error) {
		if err := r.deploymentHelper(r.deployment.Name); err != nil {
			return false, err
		}

		if r.deployment.Status.UpdatedReplicas != *r.deployment.Spec.Replicas ||
			r.deployment.Status.Replicas != *r.deployment.Spec.Replicas ||
			r.deployment.Status.AvailableReplicas != *r.deployment.Spec.Replicas {
			return false, nil
		}

		return true, nil
	})
}

func (r *Helper) removePendingPods() error {
	pods, err := r.client.CoreV1().Pods(metav1.NamespaceSystem).List(
		metav1.ListOptions{},
	)
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		ready := false
		for _, cond := range pod.Status.Conditions {
			if cond.Type == "Ready" {
				ready = true
				break
			}
		}
		if !ready {
			klog.Warningf("removing pending pod: %s", pod.Name)
			err := r.client.CoreV1().Pods(metav1.NamespaceSystem).Delete(pod.Name, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
