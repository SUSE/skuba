package caaspctl

import (
	"github.com/spf13/cobra"

	"suse.com/caaspctl/internal/app/caaspctl/cluster"
)

func NewClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Commands to handle a cluster",
	}

	cmd.AddCommand(
		cluster.NewInitCmd(),
		cluster.NewStatusCmd(),
	)

	return cmd
}
