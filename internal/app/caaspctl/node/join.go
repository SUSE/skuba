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

package node

import (
	"k8s.io/klog"

	"github.com/spf13/cobra"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments/ssh"
	node "github.com/SUSE/caaspctl/pkg/caaspctl/actions/node/join"
)

type joinOptions struct {
	target                string
	user                  string
	sudo                  bool
	port                  int
	role                  string
	ignorePreflightErrors string
}

func NewJoinCmd() *cobra.Command {
	joinOptions := joinOptions{}

	cmd := &cobra.Command{
		Use:   "join <node-name>",
		Short: "Joins a new node to the cluster",
		Run: func(cmd *cobra.Command, nodenames []string) {
			joinConfiguration := deployments.JoinConfiguration{
				KubeadmExtraArgs: map[string]string{"ignore-preflight-errors": joinOptions.ignorePreflightErrors},
			}

			switch joinOptions.role {
			case "master":
				joinConfiguration.Role = deployments.MasterRole
			case "worker":
				joinConfiguration.Role = deployments.WorkerRole
			default:
				klog.Fatalf("invalid role provided: %q, 'master' or 'worker' are the only accepted roles", joinOptions.role)
			}

			node.Join(joinConfiguration,
				ssh.NewTarget(
					nodenames[0],
					joinOptions.target,
					joinOptions.user,
					joinOptions.sudo,
					joinOptions.port,
				),
			)
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&joinOptions.target, "target", "t", "", "IP or FQDN of the node to connect to using SSH")
	cmd.Flags().StringVarP(&joinOptions.user, "user", "u", "root", "User identity used to connect to target")
	cmd.Flags().BoolVarP(&joinOptions.sudo, "sudo", "s", false, "Run remote command via sudo")
	cmd.Flags().IntVarP(&joinOptions.port, "port", "p", 22, "Port to connect to using SSH")
	cmd.Flags().StringVarP(&joinOptions.role, "role", "r", "", "Role that this node will have in the cluster (master|worker)")
	cmd.Flags().StringVar(&joinOptions.ignorePreflightErrors, "ignore-preflight-errors", "", "Comma separated list of preflight errors to ignore")

	cmd.MarkFlagRequired("target")
	cmd.MarkFlagRequired("role")

	return cmd
}
