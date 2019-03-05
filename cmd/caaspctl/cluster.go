package main

import (
	"github.com/spf13/cobra"

	"suse.com/caaspctl/cmd/caaspctl/cluster"
)

func newClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Commands to handle a cluster",
	}

	cmd.AddCommand(
		cluster.NewInitCmd(),
	)

	return cmd
}
