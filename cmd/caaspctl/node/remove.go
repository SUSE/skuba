package node

import (
	"github.com/spf13/cobra"

	"suse.com/caaspctl/pkg/caaspctl/actions/node/remove"
)

func NewRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <node-name>",
		Short: "Removes a node from the cluster",
		Run: func(cmd *cobra.Command, nodenames []string) {
			node.Remove(nodenames[0])
		},
		Args: cobra.ExactArgs(1),
	}

	return cmd
}
