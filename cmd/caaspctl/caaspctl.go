package main

import (
	"os"

	"github.com/spf13/cobra"

	"suse.com/caaspctl/internal/app/caaspctl"
)

func newRootCmd(args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use: "caaspctl",
	}

	cmd.AddCommand(
		caaspctl.NewClusterCmd(),
		caaspctl.NewNodeCmd(),
	)

	return cmd
}

func main() {
	cmd := newRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
