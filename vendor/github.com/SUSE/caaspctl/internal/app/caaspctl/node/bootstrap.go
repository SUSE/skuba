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
	"github.com/spf13/cobra"

	"k8s.io/klog"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments/ssh"
	node "github.com/SUSE/caaspctl/pkg/caaspctl/actions/node/bootstrap"
)

type bootstrapOptions struct {
	target                string
	user                  string
	sudo                  bool
	port                  int
	ignorePreflightErrors string
}

func NewBootstrapCmd() *cobra.Command {
	bootstrapOptions := bootstrapOptions{}

	cmd := cobra.Command{
		Use:   "bootstrap <node-name>",
		Short: "Bootstraps the first master node of the cluster",
		Run: func(cmd *cobra.Command, nodenames []string) {
			bootstrapConfiguration := deployments.BootstrapConfiguration{
				KubeadmExtraArgs: map[string]string{"ignore-preflight-errors": bootstrapOptions.ignorePreflightErrors},
			}

			err := node.Bootstrap(bootstrapConfiguration,
				ssh.NewTarget(
					nodenames[0],
					bootstrapOptions.target,
					bootstrapOptions.user,
					bootstrapOptions.sudo,
					bootstrapOptions.port,
				),
			)
			if err != nil {
				klog.Fatal(err)
			}
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&bootstrapOptions.target, "target", "t", "", "IP or FQDN of the node to connect to using SSH")
	cmd.Flags().StringVarP(&bootstrapOptions.user, "user", "u", "root", "User identity used to connect to target")
	cmd.Flags().IntVarP(&bootstrapOptions.port, "port", "p", 22, "Port to connect to using SSH")
	cmd.Flags().BoolVarP(&bootstrapOptions.sudo, "sudo", "s", false, "Run remote command via sudo")
	cmd.Flags().StringVar(&bootstrapOptions.ignorePreflightErrors, "ignore-preflight-errors", "", "Comma separated list of preflight errors to ignore")

	cmd.MarkFlagRequired("target")

	return &cmd
}
