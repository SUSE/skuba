package main

import (
	"log"

	"github.com/spf13/cobra"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/ssh"
	"suse.com/caaspctl/pkg/caaspctl/actions/bootstrap"
)

func newBootstrapCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "bootstrap <target>",
		Short: "Bootstraps the first master node of the cluster",
		Run: func(cmd *cobra.Command, targets []string) {
			user, err := cmd.Flags().GetString("user")
			if err != nil {
				log.Fatalf("Unable to parse user flag: %v", err)
			}
			sudo, err := cmd.Flags().GetBool("sudo")
			if err != nil {
				log.Fatalf("Unable to parse sudo flag: %v", err)
			}

			err = bootstrap.Bootstrap(ssh.NewTarget(targets[0], user, sudo))
			if err != nil {
				log.Fatal(err)
			}
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("user", "u", "root", "User identity used to connect to target")
	cmd.Flags().Bool("sudo", false, "Run remote command via sudo")

	return &cmd
}
