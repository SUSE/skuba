package main

import (
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd(args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use: "caaspctl",
	}

	cmd.AddCommand(
		newInitCmd(),
		newBootstrapCmd(),
		newJoinCmd(),
		newDeleteNodeCmd(),
	)

	return cmd
}

func main() {
	cmd := newRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
