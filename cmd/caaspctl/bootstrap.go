package main

import (
	"github.com/spf13/cobra"

	"suse.com/caaspctl/pkg/caaspctl/actions/bootstrap"
)

func newBootstrapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "bootstraps the first master node of the cluster",
		Run: func(cmd *cobra.Command, targets []string) {
			bootstrap.Bootstrap(targets[0])
		},
		Args: cobra.ExactArgs(1),
	}
}
