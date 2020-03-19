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

package upgrade

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	"github.com/SUSE/skuba/internal/pkg/skuba/testutil"

	"github.com/pmezard/go-difflib/difflib"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes/fake"
	kubeadmapiv1beta2 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	configutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"
	"sigs.k8s.io/yaml"
)

func TestPlan(t *testing.T) {
	scenarios := []struct {
		name                       string
		controlPlaneNodes          []*v1.Node
		workerNodes                []*v1.Node
		currentAddons              *kubernetes.AddonsVersion
		clusterAddonsKnownVersions map[string]*kubernetes.AddonsVersion
		currentClusterVersion      string
		availableVersions          []string
		expectedOutput             string
		expectedErr                error
	}{
		{
			name:              "Everything is up to date",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       workerNodes("1.16.0"),
			currentAddons:     updatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.16.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.16.0

All nodes match the current cluster version: 1.16.0.

Addons at the current cluster version 1.16.0 are up to date.
`,
		},
		{
			name:              "Platform up to date, drifted worker (does not tolerate cluster version)",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       workerNodes("1.14.0"),
			currentAddons:     updatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.16.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.16.0

Some nodes do not match the current cluster version (1.16.0):
  - control-plane-0: up to date
  - worker-0; current version: 1.14.0 (upgrade required)

Addons at the current cluster version 1.16.0 are up to date.
`,
		},
		{
			name:              "Platform up to date, drifted unschedulable worker (does not tolerate cluster version)",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       unschedulableWorkerNodes("1.14.0"),
			currentAddons:     updatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.16.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.16.0

Some nodes do not match the current cluster version (1.16.0):
  - control-plane-0: up to date
  - worker-unschedulable-0; current version: 1.14.0 (upgrade required); unschedulable, ignored

Addons at the current cluster version 1.16.0 are up to date.
`,
		},
		{
			name:              "Platform up to date, slightly drifted worker (tolerates cluster version)",
			controlPlaneNodes: controlPlaneNodes("1.16.1"),
			workerNodes:       workerNodes("1.16.0"),
			currentAddons:     updatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
				"1.16.1": updatedAddonsVersion(),
			},
			currentClusterVersion: "1.16.1",
			availableVersions:     []string{"1.16.0", "1.16.1"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.1
Latest Kubernetes version: 1.16.1

Some nodes do not match the current cluster version (1.16.1):
  - control-plane-0: up to date
  - worker-0; current version: 1.16.0 (upgrade suggested)

Addons at the current cluster version 1.16.1 are up to date.
`,
		},
		{
			name:              "Platform up to date, drifted worker (tolerates cluster version)",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       workerNodes("1.15.0"),
			currentAddons:     updatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.15.0": updatedAddonsVersion(),
				"1.16.0": updatedAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.15.0", "1.16.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.16.0

Some nodes do not match the current cluster version (1.16.0):
  - control-plane-0: up to date
  - worker-0; current version: 1.15.0 (upgrade suggested)

Addons at the current cluster version 1.16.0 are up to date.
`,
		},
		{
			name:              "Platform up to date, addon upgrade available for current cluster version",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       workerNodes("1.16.0"),
			currentAddons:     outdatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.16.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.16.0

All nodes match the current cluster version: 1.16.0.

Addon upgrades for 1.16.0:
  - cilium: 1.0.0 -> 1.0.1
  - dex: 1.0.0 -> 1.0.1
  - gangway: 1.0.0 -> 1.0.1
  - kured: 1.0.0 -> 1.0.1
  - psp (manifest version from 1 to 2)
`,
		},
		{
			name:              "Platform upgrade available, addon upgrade available for current cluster version",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       workerNodes("1.16.0"),
			currentAddons:     outdatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
				"1.17.0": updatedAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.16.0", "1.17.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.17.0

All nodes match the current cluster version: 1.16.0.

Addon upgrades for 1.16.0:
  - cilium: 1.0.0 -> 1.0.1
  - dex: 1.0.0 -> 1.0.1
  - gangway: 1.0.0 -> 1.0.1
  - kured: 1.0.0 -> 1.0.1
  - psp (manifest version from 1 to 2)

It is required to run 'skuba addon upgrade apply' before starting the platform upgrade.

Upgrade path to update from 1.16.0 to 1.17.0:
  - 1.16.0 -> 1.17.0

There is no need to run 'skuba addon upgrade apply' after you have completed the platform upgrade.
`,
		},
		{
			name:              "Platform upgrade available, addon upgrade available for current cluster version and for target cluster version",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       workerNodes("1.16.0"),
			currentAddons:     outdatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
				"1.17.0": latestAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.16.0", "1.17.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.17.0

All nodes match the current cluster version: 1.16.0.

Addon upgrades for 1.16.0:
  - cilium: 1.0.0 -> 1.0.1
  - dex: 1.0.0 -> 1.0.1
  - gangway: 1.0.0 -> 1.0.1
  - kured: 1.0.0 -> 1.0.1
  - psp (manifest version from 1 to 2)

It is required to run 'skuba addon upgrade apply' before starting the platform upgrade.

Upgrade path to update from 1.16.0 to 1.17.0:
  - 1.16.0 -> 1.17.0

Addon upgrades from 1.16.0 to 1.17.0:
  - cilium: 1.0.1 -> 1.0.2
  - dex: 1.0.1 -> 1.0.2
  - gangway: 1.0.1 -> 1.0.2
  - kured: 1.0.1 -> 1.0.2
  - psp (manifest version from 2 to 3)

It is required to run 'skuba addon upgrade apply' after you have completed the platform upgrade.
`,
		},
		{
			name:              "Platform upgrade available, addon upgrade available for current cluster version and for target cluster version. Control plane node updated",
			controlPlaneNodes: controlPlaneNodes("1.17.0"),
			workerNodes:       workerNodes("1.16.0"),
			currentAddons:     updatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
				"1.17.0": latestAddonsVersion(),
			},
			currentClusterVersion: "1.17.0",
			availableVersions:     []string{"1.16.0", "1.17.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.17.0
Latest Kubernetes version: 1.17.0

Some nodes do not match the current cluster version (1.17.0):
  - control-plane-0: up to date
  - worker-0; current version: 1.16.0 (upgrade suggested)

Addon upgrades for 1.17.0:
  - cilium: 1.0.1 -> 1.0.2
  - dex: 1.0.1 -> 1.0.2
  - gangway: 1.0.1 -> 1.0.2
  - kured: 1.0.1 -> 1.0.2
  - psp (manifest version from 2 to 3)
`,
		},
		{
			name:              "Several platform upgrades available, addon upgrade available for current cluster version and for target cluster version",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       workerNodes("1.16.0"),
			currentAddons:     outdatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
				"1.17.0": latestAddonsVersion(),
				"1.18.0": latestAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.16.0", "1.17.0", "1.18.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.18.0

All nodes match the current cluster version: 1.16.0.

Addon upgrades for 1.16.0:
  - cilium: 1.0.0 -> 1.0.1
  - dex: 1.0.0 -> 1.0.1
  - gangway: 1.0.0 -> 1.0.1
  - kured: 1.0.0 -> 1.0.1
  - psp (manifest version from 1 to 2)

It is required to run 'skuba addon upgrade apply' before starting the platform upgrade.

Upgrade path to update from 1.16.0 to 1.18.0:
  - 1.16.0 -> 1.17.0
  - 1.17.0 -> 1.18.0

Addon upgrades from 1.16.0 to 1.17.0:
  - cilium: 1.0.1 -> 1.0.2
  - dex: 1.0.1 -> 1.0.2
  - gangway: 1.0.1 -> 1.0.2
  - kured: 1.0.1 -> 1.0.2
  - psp (manifest version from 2 to 3)

It is required to run 'skuba addon upgrade apply' after you have completed the platform upgrade.
`,
		},
		{
			name:              "Platform upgrade available. Addons up to date for the current cluster version. Addon upgrade available for the target cluster version",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       workerNodes("1.16.0"),
			currentAddons:     updatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.16.0": updatedAddonsVersion(),
				"1.17.0": latestAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.16.0", "1.17.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.17.0

All nodes match the current cluster version: 1.16.0.

Addons at the current cluster version 1.16.0 are up to date.
There is no need to run 'skuba addon upgrade apply' before starting the platform upgrade.

Upgrade path to update from 1.16.0 to 1.17.0:
  - 1.16.0 -> 1.17.0

Addon upgrades from 1.16.0 to 1.17.0:
  - cilium: 1.0.1 -> 1.0.2
  - dex: 1.0.1 -> 1.0.2
  - gangway: 1.0.1 -> 1.0.2
  - kured: 1.0.1 -> 1.0.2
  - psp (manifest version from 2 to 3)

It is required to run 'skuba addon upgrade apply' after you have completed the platform upgrade.
`,
		},
		{
			name:              "Platform upgrade available. Does not know how to infer the upgrade",
			controlPlaneNodes: controlPlaneNodes("1.16.0"),
			workerNodes:       workerNodes("1.16.0"),
			currentAddons:     updatedAddonsVersion(),
			clusterAddonsKnownVersions: map[string]*kubernetes.AddonsVersion{
				"1.17.0": latestAddonsVersion(),
			},
			currentClusterVersion: "1.16.0",
			availableVersions:     []string{"1.18.0"},
			expectedOutput: `Current Kubernetes cluster version: 1.16.0
Latest Kubernetes version: 1.18.0
`,
			expectedErr: errors.New("cannot infer how to upgrade from 1.16.0 to 1.18.0"),
		},
	}
	for _, tt := range scenarios {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			allResources := []runtime.Object{}
			for _, node := range tt.controlPlaneNodes {
				allResources = append(allResources, node)
			}
			for _, pod := range withControlPlaneComponents(tt.controlPlaneNodes) {
				allResources = append(allResources, pod)
			}
			for _, node := range tt.workerNodes {
				allResources = append(allResources, node)
			}
			if tt.currentAddons != nil {
				allResources = append(allResources, skubaConfigMap(tt.currentAddons))
			}
			allResources = append(allResources, kubeadmConfigMap(tt.currentClusterVersion))
			clientset := fake.NewSimpleClientset(allResources...)
			availableVersions := []*version.Version{}
			for _, availableVersion := range tt.availableVersions {
				availableVersions = append(availableVersions, version.MustParseSemantic(availableVersion))
			}
			planOutput := captureOutput(func() {
				err := plan(clientset, availableVersions, func(clusterVersion *version.Version) kubernetes.AddonsVersion {
					return *tt.clusterAddonsKnownVersions[clusterVersion.String()]
				})
				if err != nil && tt.expectedErr == nil {
					t.Errorf("received error: '%v', but was not expecting an error", err)
				} else if err == nil && tt.expectedErr != nil {
					t.Errorf("did not receive an error, but was expecting: '%v'", tt.expectedErr)
				} else if err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
					t.Errorf("received error: '%v', but was expecting error: '%v'", err, tt.expectedErr)
				}
			})
			if planOutput != tt.expectedOutput {
				diff := difflib.UnifiedDiff{
					A:        difflib.SplitLines(tt.expectedOutput),
					B:        difflib.SplitLines(planOutput),
					FromFile: "Expected",
					ToFile:   "Got",
					Context:  3,
				}
				unifiedDiff, _ := difflib.GetUnifiedDiffString(diff)
				t.Errorf("plan output does not match expectation:\n%+v", unifiedDiff)
			}
		})
	}
}

func captureOutput(f func()) string {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	return string(out)
}

func skubaConfigMap(addonsVersion *kubernetes.AddonsVersion) *v1.ConfigMap {
	skubaConfigContents, _ := yaml.Marshal(&skuba.SkubaConfiguration{
		AddonsVersion: *addonsVersion,
	})
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      skuba.ConfigMapName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			skuba.SkubaConfigurationKeyName: string(skubaConfigContents),
		},
	}
}

func kubeadmConfigMap(currentClusterVersion string) *v1.ConfigMap {
	kubeadmConfigContents, _ := configutil.MarshalKubeadmConfigObject(&kubeadmapiv1beta2.ClusterConfiguration{
		KubernetesVersion: currentClusterVersion,
	})
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeadmconstants.KubeadmConfigConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			kubeadmconstants.ClusterConfigurationConfigMapKey: string(kubeadmConfigContents),
		},
	}
}

func controlPlaneNodes(versions ...string) []*v1.Node {
	ret := []*v1.Node{}
	for i, version := range versions {
		node := testutil.ControlPlaneNode(fmt.Sprintf("control-plane-%d", i))
		node.Status.NodeInfo.KubeletVersion = version
		node.Status.NodeInfo.ContainerRuntimeVersion = fmt.Sprintf("container-runtime-engine://%s", version)
		ret = append(ret, node)
	}
	return ret
}

func workerNodes(versions ...string) []*v1.Node {
	ret := []*v1.Node{}
	for i, version := range versions {
		node := testutil.WorkerNode(fmt.Sprintf("worker-%d", i))
		node.Status.NodeInfo.KubeletVersion = version
		node.Status.NodeInfo.ContainerRuntimeVersion = fmt.Sprintf("container-runtime-engine://%s", version)
		ret = append(ret, node)
	}
	return ret
}

func unschedulableWorkerNodes(versions ...string) []*v1.Node {
	ret := []*v1.Node{}
	for i, version := range versions {
		node := testutil.WorkerNode(fmt.Sprintf("worker-unschedulable-%d", i))
		node.Spec.Unschedulable = true
		node.Status.NodeInfo.KubeletVersion = version
		node.Status.NodeInfo.ContainerRuntimeVersion = fmt.Sprintf("container-runtime-engine://%s", version)
		ret = append(ret, node)
	}
	return ret
}

func withControlPlaneComponents(nodes []*v1.Node) []*v1.Pod {
	ret := []*v1.Pod{}
	components := []string{"kube-apiserver", "kube-controller-manager", "kube-scheduler", "etcd"}
	for _, node := range nodes {
		for _, component := range components {
			ret = append(ret, &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s", component, node.ObjectMeta.Name),
					Namespace: metav1.NamespaceSystem,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: fmt.Sprintf("%s-%s", component, node.ObjectMeta.Name),
							// For simplicity, image tag is set to the kubelet version. If a node is outdated,
							// all their components are. If it's up to date, all its components are as well.
							Image: fmt.Sprintf("some-registry/some-namespace/%s:%s", component, node.Status.NodeInfo.KubeletVersion),
						},
					},
				},
			})
		}
	}
	return ret
}

func addonVersion(version string, manifestVersion uint) *kubernetes.AddonVersion {
	return &kubernetes.AddonVersion{
		Version:         version,
		ManifestVersion: manifestVersion,
	}
}

func outdatedAddonsVersion() *kubernetes.AddonsVersion {
	return &kubernetes.AddonsVersion{
		kubernetes.Cilium:  addonVersion("1.0.0", 1),
		kubernetes.Kured:   addonVersion("1.0.0", 1),
		kubernetes.Dex:     addonVersion("1.0.0", 1),
		kubernetes.Gangway: addonVersion("1.0.0", 1),
		kubernetes.PSP:     addonVersion("", 1),
	}
}

func updatedAddonsVersion() *kubernetes.AddonsVersion {
	return &kubernetes.AddonsVersion{
		kubernetes.Cilium:  addonVersion("1.0.1", 2),
		kubernetes.Kured:   addonVersion("1.0.1", 2),
		kubernetes.Dex:     addonVersion("1.0.1", 2),
		kubernetes.Gangway: addonVersion("1.0.1", 2),
		kubernetes.PSP:     addonVersion("", 2),
	}
}

func latestAddonsVersion() *kubernetes.AddonsVersion {
	return &kubernetes.AddonsVersion{
		kubernetes.Cilium:  addonVersion("1.0.2", 3),
		kubernetes.Kured:   addonVersion("1.0.2", 3),
		kubernetes.Dex:     addonVersion("1.0.2", 3),
		kubernetes.Gangway: addonVersion("1.0.2", 3),
		kubernetes.PSP:     addonVersion("", 3),
	}
}
