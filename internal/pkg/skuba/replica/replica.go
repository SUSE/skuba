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

package replica

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

type affinity int

const (
	preferred affinity = iota
	required
)

const (
	highAvailabilitylabel = "caasp.suse.com/skuba-replica-ha"
	minSize               = 2
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
	Client            clientset.Interface
	ClusterSize       int
	SelectDeployments appsv1.DeploymentList
	Deployment        appsv1.Deployment
	ReplicaSize       int
	MinSize           int
}

// NewHelper creates a helper to update deployment replicas.
func NewHelper(client clientset.Interface) (*Helper, error) {
	fromSelectors, err := client.AppsV1().Deployments(metav1.NamespaceSystem).List(
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=true", highAvailabilitylabel),
		},
	)
	if err != nil {
		return nil, err
	}

	node, err := kubernetes.GetAllNodes(client)
	if err != nil {
		return nil, err
	}

	return &Helper{
		Client:            client,
		ClusterSize:       len(node.Items),
		MinSize:           minSize,
		SelectDeployments: *fromSelectors,
	}, nil
}

func (r *Helper) replaceAffinity(remove affinity, create affinity) (bool, error) {
	// doRemove will remove only when affinity path exists
	doRemove := false
	affinity := r.Deployment.Spec.Template.Spec.Affinity
	if affinity != nil && affinity.PodAntiAffinity != nil {
		preferredList := len(r.Deployment.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
		requiredList := len(r.Deployment.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
		switch remove {
		case preferred:
			if preferredList != 0 {
				doRemove = true
				break
			}
			if requiredList != 0 {
				return false, nil
			}
		case required:
			if requiredList != 0 {
				doRemove = true
				break
			}
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
		affinityJSON = fmt.Sprintf(patchAffinityRequired, r.Deployment.ObjectMeta.Name)
	default:
		affinityJSON = fmt.Sprintf(patchAffinityPreferred, r.Deployment.ObjectMeta.Name)
	}

	_, err := r.Client.AppsV1().Deployments(metav1.NamespaceSystem).Patch(r.Deployment.ObjectMeta.Name, types.StrategicMergePatchType, []byte(affinityJSON))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *Helper) removeAffinity() error {
	_, err := r.Client.AppsV1().Deployments(metav1.NamespaceSystem).Patch(r.Deployment.ObjectMeta.Name, types.JSONPatchType, []byte(patchAffinityRemove))
	if err != nil {
		return err
	}
	return nil
}

func (r *Helper) updateDeploymentReplica(size int) (*appsv1.Deployment, error) {
	replicaJSON := fmt.Sprintf(patchReplicas, size)
	return r.Client.AppsV1().Deployments(metav1.NamespaceSystem).Patch(r.Deployment.ObjectMeta.Name, types.StrategicMergePatchType, []byte(replicaJSON))
}

func (r *Helper) refreshReplicas() error {
	// At least one node needs to be free of the affinity rules for pods to re-distribute.
	// Sizing down replicas would make room for new pod resource.
	// For example, during rolling upgrade if all node already has `required`, scheduler will `Pend`
	// when tries to create a new deployment with `preferred`. This is due to the conflict to its original rule.
	// But this ends up with less of replicas. So need the following patch to bring the replicas to the right size.
	if _, err := r.updateDeploymentReplica(r.ClusterSize - 1); err != nil {
		return err
	}
	if _, err := r.updateDeploymentReplica(r.ReplicaSize); err != nil {
		return err
	}
	return nil
}

// UpdateNodes patches for replicas affinity rules after adding of nodes.
// This updates affinity rule to preferredDuringSchedulingIgnoredDuringExecution of
// cluster size less then replica size,
// And updates affinity rule to requiredDuringSchedulingIgnoredDuringExecution when
// cluster size is equal or greater to replicas.
func (r *Helper) UpdateNodes() error {
	for _, deployment := range r.SelectDeployments.Items {
		r.Deployment = deployment
		r.ReplicaSize = int(*r.Deployment.Spec.Replicas)
		if r.ReplicaSize == 0 {
			r.ReplicaSize = r.MinSize
		}

		nodes := r.ClusterSize
		isAffinityReplaced := false
		var err error
		switch {
		case nodes >= r.ReplicaSize:
			isAffinityReplaced, err = r.replaceAffinity(preferred, required)
			if err != nil {
				return err
			}
		case nodes >= r.MinSize:
			isAffinityReplaced, err = r.replaceAffinity(required, preferred)
			if err != nil {
				return err
			}
		}
		if !isAffinityReplaced && nodes <= r.ReplicaSize {
			if err := r.refreshReplicas(); err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateDrainNodes patches for replicas affinity rules before removing of nodes.
// This updates affinity rule to preferredDuringSchedulingIgnoredDuringExecution when
// cluster size equal or less then replicas.
func (r *Helper) UpdateDrainNodes() error {
	for _, deployment := range r.SelectDeployments.Items {
		r.Deployment = deployment
		r.ReplicaSize = int(*r.Deployment.Spec.Replicas)
		if r.ClusterSize > r.ReplicaSize {
			continue
		}

		if _, err := r.replaceAffinity(required, preferred); err != nil {
			return err
		}
		if err := r.refreshReplicas(); err != nil {
			return err
		}
	}
	return nil
}
