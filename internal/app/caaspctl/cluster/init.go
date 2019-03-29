/*
 * Copyright (c) 2019 SUSE LLC. All rights reserved.
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
	"github.com/spf13/cobra"

	cluster "suse.com/caaspctl/pkg/caaspctl/actions/cluster/init"
)

type initOptions struct {
	ControlPlane string
}

func NewInitCmd() *cobra.Command {
	initOptions := initOptions{}

	cmd := &cobra.Command{
		Use:   "init <cluster-name>",
		Short: "Initialize caaspctl structure for cluster deployment",
		Run: func(cmd *cobra.Command, args []string) {
			cluster.Init(cluster.InitConfiguration{
				ClusterName:  args[0],
				ControlPlane: initOptions.ControlPlane,
			})
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&initOptions.ControlPlane, "control-plane", "", "The control plane location that will load balance the master nodes")
	cmd.MarkFlagRequired("control-plane")

	return cmd
}
