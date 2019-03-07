package node

import (
	"log"

	"github.com/spf13/cobra"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/ssh"
	node "suse.com/caaspctl/pkg/caaspctl/actions/node/bootstrap"
)

type bootstrapOptions struct {
	target string
	user   string
	sudo   bool
	port   int
}

func NewBootstrapCmd() *cobra.Command {
	bootstrapOptions := bootstrapOptions{}

	cmd := cobra.Command{
		Use:   "bootstrap <node-name>",
		Short: "Bootstraps the first master node of the cluster",
		Run: func(cmd *cobra.Command, nodenames []string) {
			err := node.Bootstrap(
				ssh.NewTarget(
					nodenames[0],
					bootstrapOptions.target,
					bootstrapOptions.user,
					bootstrapOptions.sudo,
					bootstrapOptions.port,
				),
			)
			if err != nil {
				log.Fatal(err)
			}
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVarP(&bootstrapOptions.target, "target", "t", "", "IP or FQDN of the node to connect to using SSH")
	cmd.Flags().StringVarP(&bootstrapOptions.user, "user", "u", "root", "User identity used to connect to target")
	cmd.Flags().IntVarP(&bootstrapOptions.port, "port", "p", 22, "Port to connect to using SSH")
	cmd.Flags().BoolVarP(&bootstrapOptions.sudo, "sudo", "s", false, "Run remote command via sudo")

	cmd.MarkFlagRequired("target")

	return &cmd
}
