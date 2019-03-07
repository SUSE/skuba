package node

import (
	"log"

	"github.com/spf13/cobra"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/ssh"
	node "suse.com/caaspctl/pkg/caaspctl/actions/node/join"
)

type joinOptions struct {
	target string
	user   string
	sudo   bool
	port   int
	role   string
}

func NewJoinCmd() *cobra.Command {
	joinOptions := joinOptions{}

	cmd := &cobra.Command{
		Use:   "join <node-name>",
		Short: "Joins a new node to the cluster",
		Run: func(cmd *cobra.Command, nodenames []string) {
			joinConfiguration := deployments.JoinConfiguration{}

			switch joinOptions.role {
			case "master":
				joinConfiguration.Role = deployments.MasterRole
			case "worker":
				joinConfiguration.Role = deployments.WorkerRole
			default:
				log.Fatalf("invalid role provided: %q, 'master' or 'worker' are the only accepted roles", joinOptions.role)
			}

			node.Join(joinConfiguration,
				ssh.NewTarget(
					nodenames[0],
					joinOptions.target,
					joinOptions.user,
					joinOptions.sudo,
					joinOptions.port,
				),
			)
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&joinOptions.target, "target", "t", "", "IP or FQDN of the node to connect to using SSH")
	cmd.Flags().StringVarP(&joinOptions.user, "user", "u", "root", "User identity used to connect to target")
	cmd.Flags().BoolVarP(&joinOptions.sudo, "sudo", "s", false, "Run remote command via sudo")
	cmd.Flags().IntVarP(&joinOptions.port, "port", "p", 22, "Port to connect to using SSH")
	cmd.Flags().StringVarP(&joinOptions.role, "role", "r", "", "Role that this node will have in the cluster (master|worker)")

	cmd.MarkFlagRequired("target")
	cmd.MarkFlagRequired("role")

	return cmd
}
