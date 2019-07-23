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

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog"

	cilium "github.com/SUSE/skuba/internal/pkg/skuba/cni"
	"github.com/SUSE/skuba/internal/pkg/skuba/dex"
	"github.com/SUSE/skuba/internal/pkg/skuba/gangway"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/kured"
	"github.com/SUSE/skuba/pkg/skuba"
	cluster "github.com/SUSE/skuba/pkg/skuba/actions/cluster/init"
)

type initOptions struct {
	ControlPlane      string
	KubernetesVersion string
	CloudProvider     string
}

// NewInitCmd creates a new `skuba cluster init` cobra command
func NewInitCmd() *cobra.Command {
	initOptions := initOptions{}

	cmd := &cobra.Command{
		Use:   "init <cluster-name> --control-plane <IP/FQDN>",
		Short: "Initialize skuba structure for cluster deployment",
		Run: func(cmd *cobra.Command, args []string) {
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
				ClusterName:         args[0],
				CloudProvider:       initOptions.CloudProvider,
				ControlPlane:        initOptions.ControlPlane,
				CiliumImage:         cilium.GetCiliumImage(),
				CiliumInitImage:     cilium.GetCiliumInitImage(),
				CiliumOperatorImage: cilium.GetCiliumOperatorImage(),
				KuredImage:          kured.GetKuredImage(),
				DexImage:            dex.GetDexImage(),
				GangwayClientSecret: dex.GenerateClientSecret(),
				GangwayImage:        gangway.GetGangwayImage(),
				KubernetesVersion:   kubernetesVersion.String(),
				ImageRepository:     skuba.ImageRepository,
				EtcdImageTag:        kubernetes.ComponentVersionForClusterVersion(kubernetes.Etcd, kubernetesVersion),
				CoreDNSImageTag:     kubernetes.ComponentVersionForClusterVersion(kubernetes.CoreDNS, kubernetesVersion),
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
	cmd.MarkFlagRequired("control-plane")

	return cmd
}
