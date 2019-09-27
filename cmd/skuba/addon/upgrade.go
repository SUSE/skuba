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

package addons

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	addons "github.com/SUSE/skuba/pkg/skuba/actions/addon/upgrade"
)

// NewUpgradeCmd creates a new `skuba addon upgrade` cobra command
func NewUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Manages addon upgrade operations",
	}

	cmd.AddCommand(
		newUpgradePlanCmd(),
	)

	return cmd
}

func newUpgradePlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Plan addon upgrade",
		Run: func(cmd *cobra.Command, args []string) {
			if err := addons.Plan(); err != nil {
				fmt.Printf("Unable to plan addon upgrade: %s\n", err)
				os.Exit(1)
			}
		},
		Args: cobra.NoArgs,
	}
}
