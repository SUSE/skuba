package cluster

import (
	"github.com/spf13/cobra"
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show cluster status",
		Run: func(cmd *cobra.Command, args []string) {
		},
		Args: cobra.NoArgs,
	}
}
