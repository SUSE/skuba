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
	"os"

	"github.com/spf13/cobra"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments/ssh"
	"github.com/SUSE/skuba/pkg/skuba/actions/node/upgrade"
)

func NewUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Manages node specific upgrade operations",
	}

	cmd.AddCommand(
		newUpgradePlanCmd(),
		newUpgradeApplyCmd(),
	)

	return cmd
}

func newUpgradePlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plan <node-name>",
		Short: "Plan node upgrade",
		Run: func(cmd *cobra.Command, args []string) {
			if err := upgrade.Plan(args[0]); err != nil {
				fmt.Printf("Unable to plan node upgrade: %s\n", err)
				os.Exit(1)
			}
		},
		Args: cobra.ExactArgs(1),
	}
}

func newUpgradeApplyCmd() *cobra.Command {
	target := ssh.Target{}
	cmd := cobra.Command{
		Use:   "apply",
		Short: "Apply node upgrade",
		Run: func(cmd *cobra.Command, args []string) {
			if err := upgrade.Apply(target.GetDeployment("")); err != nil {
				fmt.Printf("Unable to apply node upgrade: %s\n", err)
				os.Exit(1)
			}
		},
		Args: cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(target.GetFlags())
	return &cmd
}
