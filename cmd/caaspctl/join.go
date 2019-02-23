package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
	"suse.com/caaspctl/pkg/caaspctl/actions/join"
)

type JoinOptions struct {
	Role string
}

func newJoinCmd() *cobra.Command {
	joinOptions := JoinOptions{}

	cmd := &cobra.Command{
		Use:   "join <target>",
		Short: "Joins a new node to the cluster",
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

			target := salt.Target{
				Node: targets[0],
				User: user,
				Sudo: sudo,
			}

			joinConfiguration := join.JoinConfiguration{
				Target: target,
			}

			switch joinOptions.Role {
			case "master":
				joinConfiguration.Role = join.MasterRole
			case "worker":
				joinConfiguration.Role = join.WorkerRole
			default:
				log.Fatalf("Invalid role provided: %q, 'master' or 'worker' are the only accepted roles", joinOptions.Role)
			}

			masterConfig := salt.NewMasterConfig(
				saltPath,
			)
			defer os.RemoveAll(masterConfig.GetTempDir(target))

			join.Join(
				joinConfiguration,
				masterConfig,
			)
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&joinOptions.Role, "role", "", "", "Role that this node will have in the cluster (master|worker)")
	cmd.MarkFlagRequired("role")

	cmd.Flags().StringP("user", "u", "root", "User identity used to connect to target")
	cmd.Flags().Bool("sudo", false, "Run remote command via sudo")

	cmd.Flags().StringP("salt-path", "s", "", "Salt root path to the states folder")
	cmd.MarkFlagRequired("salt-path")

	return cmd
}
