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
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/deployments/ssh"
	"github.com/SUSE/skuba/pkg/skuba/actions"
	node "github.com/SUSE/skuba/pkg/skuba/actions/node/reset"
)

type resetOptions struct {
	ignorePreflightErrors string
}

// NewResetCmd creates a new cobra reset command
func NewResetCmd() *cobra.Command {
	resetOptions := resetOptions{}
	target := ssh.Target{}

	cmd := cobra.Command{
		Use:   "reset",
		Short: "Resets the node to it's state prior to running join or bootstrap",
		Run: func(cmd *cobra.Command, args []string) {
			resetConfiguration := deployments.ResetConfiguration{
				KubeadmExtraArgs: map[string]string{"ignore-preflight-errors": resetOptions.ignorePreflightErrors},
			}

			if err := node.Reset(resetConfiguration, target.GetDeployment("")); err != nil {
				klog.Fatalf("error resetting node: %s", err)
			}

			fmt.Println("successfully reset node to state prior to running bootstrap or join")

		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().AddFlagSet(target.GetFlags())
	actions.AddCommonFlags(&cmd, &resetOptions.ignorePreflightErrors)

	return &cmd
}
