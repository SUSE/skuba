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

package node

import (
	"github.com/spf13/cobra"

	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/deployments/ssh"
	node "github.com/SUSE/skuba/pkg/skuba/actions/node/bootstrap"
)

type bootstrapOptions struct {
	ignorePreflightErrors string
}

func NewBootstrapCmd() *cobra.Command {
	bootstrapOptions := bootstrapOptions{}
	target := ssh.Target{}

	cmd := cobra.Command{
		Use:   "bootstrap <node-name>",
		Short: "Bootstraps the first master node of the cluster",
		Run: func(cmd *cobra.Command, nodenames []string) {
			bootstrapConfiguration := deployments.BootstrapConfiguration{
				KubeadmExtraArgs: map[string]string{"ignore-preflight-errors": bootstrapOptions.ignorePreflightErrors},
			}

			d := target.GetDeployment(nodenames[0])
			if err := node.Bootstrap(bootstrapConfiguration, d); err != nil {
				klog.Fatalf("error bootstraping node: %s", err)
			}
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().AddFlagSet(target.GetFlags())
	cmd.Flags().StringVar(&bootstrapOptions.ignorePreflightErrors, "ignore-preflight-errors", "", "Comma separated list of preflight errors to ignore")

	return &cmd
}
