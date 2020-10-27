/*
 * Copyright (c) 2020 SUSE LLC.
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
	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	addons "github.com/SUSE/skuba/pkg/skuba/actions/addon/refresh"
)

// NewRefreshCmd creates a new `skuba addon refresh` cobra command
func NewRefreshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Manages addon refresh operations",
	}

	cmd.AddCommand(
		newRefreshLocalConfigCmd(),
	)

	return cmd
}

func newRefreshLocalConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "localconfig",
		Short: "Update local cluster definition folder configuration",
		Run: func(cmd *cobra.Command, args []string) {
			clientSet, err := kubernetes.GetAdminClientSet()
			if err != nil {
				klog.Errorf("unable to get admin client set: %s", err)
				os.Exit(1)
			}
			if err := addons.AddonsBaseManifest(clientSet); err != nil {
				fmt.Printf("Unable to update addons base manifests: %s\n", err)
				os.Exit(1)
			}
		},
		Args: cobra.NoArgs,
	}
}
