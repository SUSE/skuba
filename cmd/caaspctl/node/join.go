package node

import (
	"log"

	"github.com/spf13/cobra"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/ssh"
	node "suse.com/caaspctl/pkg/caaspctl/actions/node/join"
)

type JoinOptions struct {
	Role string
}

func NewJoinCmd() *cobra.Command {
	joinOptions := JoinOptions{}

	cmd := &cobra.Command{
		Use:   "join <node-name>",
		Short: "Joins a new node to the cluster",
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

			joinConfiguration := deployments.JoinConfiguration{}

			switch joinOptions.Role {
			case "master":
				joinConfiguration.Role = deployments.MasterRole
			case "worker":
				joinConfiguration.Role = deployments.WorkerRole
			default:
				log.Fatalf("Invalid role provided: %q, 'master' or 'worker' are the only accepted roles", joinOptions.Role)
			}

			node.Join(joinConfiguration, ssh.NewTarget(nodenames[0], target, user, sudo, port))
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&joinOptions.Role, "role", "", "", "Role that this node will have in the cluster (master|worker)")
	cmd.MarkFlagRequired("role")

	cmd.Flags().StringP("target", "t", "", "IP or FQDN of the node to connect to using SSH")
	cmd.MarkFlagRequired("target")

	cmd.Flags().StringP("user", "u", "root", "User identity used to connect to target")
	cmd.Flags().Bool("sudo", false, "Run remote command via sudo")

	return cmd
}
