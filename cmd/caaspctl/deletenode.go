package main

import (
	"github.com/spf13/cobra"

	"suse.com/caaspctl/pkg/caaspctl/actions/deletenode"
)

func newDeleteNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-node <target>",
		Short: "Delete a node from the cluster",
		Run: func(cmd *cobra.Command, targets []string) {
			deletenode.DeleteNode(targets[0])
		},
		Args: cobra.ExactArgs(1),
	}

	return cmd
}
