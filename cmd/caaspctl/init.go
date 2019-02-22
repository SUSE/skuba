package main

import (
	"github.com/spf13/cobra"

	doinit "suse.com/caaspctl/pkg/caaspctl/actions/init"
)

type InitOptions struct {
	ProjectName  string
	ControlPlane string
}

func newInitCmd() *cobra.Command {
	initOptions := InitOptions{}

	cmd := &cobra.Command{
		Use:   "init <cluster-name>",
		Short: "Initialize caaspctl structure for cluster deployment",
		Run: func(cmd *cobra.Command, args []string) {
			doinit.Init(doinit.InitConfiguration{
				ProjectName:  args[0],
				ControlPlane: initOptions.ControlPlane,
			})
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&initOptions.ControlPlane, "control-plane", "", "", "The control plane location that will load balance the master nodes")
	cmd.MarkFlagRequired("control-plane-endpoint")

	return cmd
}
