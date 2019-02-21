package main

import (
	"log"

	"github.com/spf13/cobra"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
	"suse.com/caaspctl/pkg/caaspctl/actions/join"
)

func newJoinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "joins a new node to the cluster",
		Run: func(cmd *cobra.Command, targets []string) {
			user, err := cmd.Flags().GetString("user")
			if err != nil {
				log.Fatalf("Unable to parse user flag: %v", err)
			}
			sudo, err := cmd.Flags().GetBool("sudo")
			if err != nil {
				log.Fatalf("Unable to parse sudo flag: %v", err)
			}

			join.Join(salt.Target{
				Node: targets[0],
				User: user,
				Sudo: sudo,
			})
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("user", "u", "root", "user identity used to connect to target")
	cmd.Flags().Bool("sudo", false, "run remote command via sudo")

	return cmd
}
