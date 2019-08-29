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
	"github.com/spf13/cobra"
	"k8s.io/klog"

	cluster "github.com/SUSE/skuba/pkg/skuba/actions/cluster/status"
)

// NewStatusCmd creates a new `skuba cluster status` cobra command
func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show cluster status",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cluster.Status(); err != nil {
				klog.Errorf("unable to get cluster status: %s", err)
			}
		},
		Args: cobra.NoArgs,
	}
}
