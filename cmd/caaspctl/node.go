package main

import (
	"github.com/spf13/cobra"

	"suse.com/caaspctl/cmd/caaspctl/node"
)

func newNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Commands to handle a specific node",
	}

	cmd.AddCommand(
		node.NewBootstrapCmd(),
		node.NewJoinCmd(),
		node.NewRemoveNodeCmd(),
	)

	return cmd
}
