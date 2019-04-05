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

package completion

import (
	"github.com/spf13/cobra"
	"os"
)

func NewBashCompletion() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generates bash completion scripts",
		// FIXME: tail -n +2 can be dropped when stdout from main() is not polluted anymore
		Long: `To load caaspctl bash completion add the following to your ~/.bashrc or ~/.profile

. <(caaspctl completion | tail -n +2)`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Parent().GenBashCompletion(os.Stdout)
		},
	}
}
