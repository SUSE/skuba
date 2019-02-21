package main

import (
	"log"

	"github.com/spf13/cobra"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
	"suse.com/caaspctl/pkg/caaspctl/actions/bootstrap"
)

func newBootstrapCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "bootstrap",
		Short: "bootstraps the first master node of the cluster",
		Run: func(cmd *cobra.Command, targets []string) {
			saltPath, err := cmd.Flags().GetString("salt-path")
			if err != nil {
				log.Fatalf("Unable to parse salt flag: %v", err)
			}
			user, err := cmd.Flags().GetString("user")
			if err != nil {
				log.Fatalf("Unable to parse user flag: %v", err)
			}
			sudo, err := cmd.Flags().GetBool("sudo")
			if err != nil {
				log.Fatalf("Unable to parse sudo flag: %v", err)
			}

			bootstrap.Bootstrap(
				salt.Target{
					Node: targets[0],
					User: user,
					Sudo: sudo,
				},
				salt.NewMasterConfig(
					saltPath,
				),
			)
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("user", "u", "root", "user identity used to connect to target")
	cmd.Flags().Bool("sudo", false, "run remote command via sudo")

	return &cmd
}
