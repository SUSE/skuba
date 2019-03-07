package cluster

import (
	"github.com/spf13/cobra"

	cluster "suse.com/caaspctl/pkg/caaspctl/actions/cluster/init"
)

type initOptions struct {
	ControlPlane string
}

func NewInitCmd() *cobra.Command {
	initOptions := initOptions{}

	cmd := &cobra.Command{
		Use:   "init <cluster-name>",
		Short: "Initialize caaspctl structure for cluster deployment",
		Run: func(cmd *cobra.Command, args []string) {
			cluster.Init(cluster.InitConfiguration{
				ClusterName:  args[0],
				ControlPlane: initOptions.ControlPlane,
			})
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&initOptions.ControlPlane, "control-plane", "", "", "The control plane location that will load balance the master nodes")
	cmd.MarkFlagRequired("control-plane-endpoint")

	return cmd
}
