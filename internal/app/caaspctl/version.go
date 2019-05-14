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

package caaspctl

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/kubernetes"
	"github.com/SUSE/caaspctl/pkg/caaspctl"
)

var (
	Version   string
	BuildDate string
	Commit    string
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stderr, "caaspctl version: %s (%s) %s %s %s\n", Version, caaspctl.BuildType, Commit, BuildDate, runtime.Version())
			fmt.Fprintf(os.Stderr, "kubernetes version: %s\n", kubernetes.CurrentVersion)
		},
	}
}
