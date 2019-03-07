package caaspctl

import (
	"github.com/spf13/cobra"

	"suse.com/caaspctl/internal/app/caaspctl/node"
)

func NewNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Commands to handle a specific node",
	}

	cmd.AddCommand(
		node.NewBootstrapCmd(),
		node.NewJoinCmd(),
		node.NewRemoveCmd(),
	)

	return cmd
}
