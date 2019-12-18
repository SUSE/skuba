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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"

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

	register(cmd.PersistentFlags(), "v")

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

// register adds a flag to local that targets the Value associated with the Flag named globalName in flag.CommandLine.
func register(local *pflag.FlagSet, globalName string) {
	if f := flag.CommandLine.Lookup(globalName); f != nil {
		pflagFlag := pflag.PFlagFromGoFlag(f)
		pflagFlag.Name = "verbosity"
		pflagFlag.Usage = "log level [0-5]. 0 (Only Error and Warning) to 5 (Maximum detail)."
		local.AddFlag(pflagFlag)
	} else {
		klog.Fatalf("failed to find flag in global flagset (flag): %s", globalName)
	}
}
