package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
	"suse.com/caaspctl/pkg/caaspctl/actions/bootstrap"
)

func newBootstrapCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "bootstrap <target>",
		Short: "Bootstraps the first master node of the cluster",
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

			masterConfig := salt.NewMasterConfig(
				saltPath,
			)
			defer os.RemoveAll(masterConfig.GetTempDir(target))

			err = bootstrap.Bootstrap(
				bootstrap.BootstrapConfiguration{
					Target: target,
				},
				masterConfig,
			)
			if err != nil {
				log.Fatal(err)
			}
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("user", "u", "root", "User identity used to connect to target")
	cmd.Flags().Bool("sudo", false, "Run remote command via sudo")

	cmd.Flags().StringP("salt-path", "s", "", "Salt root path to the states folder")
	cmd.MarkFlagRequired("salt-path")

	return &cmd
}
