package main

import (
	"github.com/spf13/cobra"

	"suse.com/caaspctl/pkg/caaspctl/actions/join"
)

func newJoinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "join",
		Short: "joins a new node to the cluster",
		Run: func(cmd *cobra.Command, targets []string) {
			join.Join(targets[0])
		},
		Args: cobra.ExactArgs(1),
	}
}
