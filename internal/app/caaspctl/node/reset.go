package node

import (
	"fmt"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments"
	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments/ssh"
	node "github.com/SUSE/caaspctl/pkg/caaspctl/actions/node/reset"
	"github.com/spf13/cobra"

	"k8s.io/klog"
)

type resetOptions struct {
	target                string
	user                  string
	sudo                  bool
	port                  int
	ignorePreflightErrors string
}

// NewResetCmd creates a new cobra reset command
func NewResetCmd() *cobra.Command {
	resetOptions := resetOptions{}

	cmd := cobra.Command{
		Use:   "reset",
		Short: "Resets the node to it's state prior to running join or bootstrap",
		Run: func(cmd *cobra.Command, args []string) {
			resetConfiguration := deployments.ResetConfiguration{
				KubeadmExtraArgs: map[string]string{"ignore-preflight-errors": resetOptions.ignorePreflightErrors},
			}

			err := node.Reset(resetConfiguration,
				ssh.NewTarget(
					"",
					resetOptions.target,
					resetOptions.user,
					resetOptions.sudo,
					resetOptions.port,
				),
			)

			if err != nil {
				klog.Fatalf("error resetting node: %s", err)
			}

			fmt.Println("successfully reset node to state prior to running bootstrap or join")

		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringVarP(&resetOptions.target, "target", "t", "", "IP or FQDN of the node to connect to using SSH")
	cmd.Flags().StringVarP(&resetOptions.user, "user", "u", "root", "User identity used to connect to target")
	cmd.Flags().BoolVarP(&resetOptions.sudo, "sudo", "s", false, "Run remote command via sudo")
	cmd.Flags().IntVarP(&resetOptions.port, "port", "p", 22, "Port to connect to using SSH")
	cmd.Flags().StringVar(&resetOptions.ignorePreflightErrors, "ignore-preflight-errors", "", "Comma separated list of preflight errors to ignore")
	cmd.MarkFlagRequired("target")

	return &cmd
}
