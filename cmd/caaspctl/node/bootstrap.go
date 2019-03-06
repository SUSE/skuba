package node

import (
	"log"

	"github.com/spf13/cobra"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/ssh"
	node "suse.com/caaspctl/pkg/caaspctl/actions/node/bootstrap"
)

func NewBootstrapCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "bootstrap <node-name>",
		Short: "Bootstraps the first master node of the cluster",
		Run: func(cmd *cobra.Command, nodenames []string) {
			target, err := cmd.Flags().GetString("target")
			if err != nil {
				log.Fatalf("Unable to parse target flag: %v", err)
			}
			user, err := cmd.Flags().GetString("user")
			if err != nil {
				log.Fatalf("Unable to parse user flag: %v", err)
			}
			sudo, err := cmd.Flags().GetBool("sudo")
			if err != nil {
				log.Fatalf("Unable to parse sudo flag: %v", err)
			}
			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				port = 22
			}

			err = node.Bootstrap(ssh.NewTarget(nodenames[0], target, user, sudo, port))
			if err != nil {
				log.Fatal(err)
			}
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringP("target", "t", "", "IP or FQDN of the node to connect to using SSH")
	cmd.MarkFlagRequired("target")

	cmd.Flags().StringP("user", "u", "root", "User identity used to connect to target")
	cmd.Flags().Bool("sudo", false, "Run remote command via sudo")

	return &cmd
}
