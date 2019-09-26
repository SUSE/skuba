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

package cluster

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	cluster "github.com/SUSE/skuba/pkg/skuba/actions/cluster/init"
)

type initOptions struct {
	ControlPlane      string
	KubernetesVersion string
	CloudProvider     string
	StrictCapDefaults bool
	RegistryMirror    string
}

func ParseCrioRegistryConfiguration(config string) cluster.CrioMirrorConfiguration {
	values := strings.Split(config, ":")
	if len(values) != 3 {
		return cluster.CrioMirrorConfiguration{}
	}
	insecure, err := strconv.ParseBool(values[2])
	// Default to secure registry
	if err != nil {
		insecure = false
	}
	return cluster.CrioMirrorConfiguration{
		SourceRegistry: values[0],
		MirrorRegistry: values[1],
		Insecure:       insecure,
	}
}

// NewInitCmd creates a new `skuba cluster init` cobra command
func NewInitCmd() *cobra.Command {
	initOptions := initOptions{}

	cmd := &cobra.Command{
		Use:   "init <cluster-name> --control-plane <IP/FQDN>",
		Short: "Initialize skuba structure for cluster deployment",
		Run: func(cmd *cobra.Command, args []string) {
			registryConfiguration := ParseCrioRegistryConfiguration(initOptions.RegistryMirror)
			kubernetesVersion := kubernetes.LatestVersion()
			if initOptions.KubernetesVersion != "" {
				var err error
				kubernetesVersion, err = version.ParseSemantic(initOptions.KubernetesVersion)
				if err != nil || !kubernetes.IsVersionAvailable(kubernetesVersion) {
					fmt.Printf("Version %s does not exist or cannot be parsed.\n", initOptions.KubernetesVersion)
					os.Exit(1)
				}
			}

			initConfig := cluster.InitConfiguration{
				ClusterName:       args[0],
				CloudProvider:     initOptions.CloudProvider,
				ControlPlane:      initOptions.ControlPlane,
				PauseImage:        images.GetGenericImage(skuba.ImageRepository, "pause", kubernetes.ComponentVersionForClusterVersion(kubernetes.Pause, kubernetesVersion)),
				KubernetesVersion: kubernetesVersion,
				ImageRepository:   skuba.ImageRepository,
				EtcdImageTag:      kubernetes.ComponentVersionForClusterVersion(kubernetes.Etcd, kubernetesVersion),
				CoreDNSImageTag:   kubernetes.ComponentVersionForClusterVersion(kubernetes.CoreDNS, kubernetesVersion),
				StrictCapDefaults: initOptions.StrictCapDefaults,
				RegistryMirror:    registryConfiguration,
			}

			err := cluster.Init(initConfig)
			if err != nil {
				klog.Fatalf("init failed due to error: %s", err)
			}
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringVar(&initOptions.ControlPlane, "control-plane", "", "The control plane location (IP/FQDN) that will load balance the master nodes (required)")
	if skuba.BuildType == "development" {
		cmd.Flags().StringVar(&initOptions.KubernetesVersion, "kubernetes-version", "", "The kubernetes version to bootstrap with (only in development build)")
	}
	cmd.Flags().StringVar(&initOptions.CloudProvider, "cloud-provider", "", "Enable cloud provider integration with the chosen cloud. Valid values: openstack")
	_ = cmd.MarkFlagRequired("control-plane")

	cmd.Flags().BoolVar(&initOptions.StrictCapDefaults, "strict-capability-defaults", false, "All the containers will start with CRI-O default capabilities")

	// TODO: Might be useful to allow configuring multiple mirror - but that
	// needs a special syntax - e.g. "source.url:mirror.url"

	// As part of that micro-syntax it should then be possible to also
	// configure the security of the mirror
	cmd.Flags().StringVar(&initOptions.RegistryMirror, "registry-mirror", "", "Configure a mirror for a registry to pull container-images from. Syntax: '<source_registry>:<target_registry>:<insecure>' (e.g. 'registry.suse.com:my.registry.org:true' would setup an insecure mirror from registry.suse.com to my.registry.org)")
	return cmd
}
