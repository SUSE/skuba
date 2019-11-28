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
	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba/actions/cluster/upgrade"
)

// NewUpgradeCmd creates a new `skuba cluster upgrade` cobra command
func NewUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Manages cluster-wide upgrade operations",
	}

	cmd.AddCommand(
		newUpgradePlanCmd(),
	)

	return cmd
}

func newUpgradePlanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Plan cluster upgrade",
		Run: func(cmd *cobra.Command, args []string) {
			clientSet, err := kubernetes.GetAdminClientSet()
			if err != nil {
				klog.Errorf("unable to get admin client set: %s", err)
				os.Exit(1)
			}
			if err := upgrade.Plan(clientSet); err != nil {
				fmt.Printf("Unable to plan cluster upgrade: %s\n", err)
				os.Exit(1)
			}
		},
		Args: cobra.NoArgs,
	}
}
