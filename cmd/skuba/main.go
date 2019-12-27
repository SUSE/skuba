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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/SUSE/skuba/cmd/skuba/flags"
	skubapkg "github.com/SUSE/skuba/pkg/skuba"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		// grab the base filename if the binary file is link
		Use: filepath.Base(os.Args[0]),
	}

	cmd.AddCommand(
		NewVersionCmd(),
		NewClusterCmd(),
		NewCompletionCmd(),
		NewNodeCmd(),
		NewAuthCmd(),
		NewAddonCmd(),
	)

	flags.RegisterVerboseFlag(cmd.PersistentFlags())

	return cmd
}

func main() {
	printVersionStatement()
	klog.InitFlags(nil)
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printVersionStatement() {
	switch {
	case skubapkg.Tag == "":
		fmt.Println("** This is an UNTAGGED version and NOT intended for production usage. **")
	case strings.Contains(skubapkg.Tag, "alpha"):
		fmt.Println("** This is an ALPHA release and NOT intended for production usage. **")
	case strings.Contains(skubapkg.Tag, "beta"):
		fmt.Println("** This is a BETA release and NOT intended for production usage. **")
	case strings.Contains(skubapkg.Tag, "rc"):
		fmt.Println("** This is a RC release and NOT intended for production usage. **")
	case skubapkg.BuildType != "release":
		fmt.Printf("** This is a tagged version (%s), but is not built in release mode (mode: %q.) **\n", skubapkg.Tag, skubapkg.BuildType)
	}
}
