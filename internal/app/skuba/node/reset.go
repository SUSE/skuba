package node

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
	"github.com/SUSE/skuba/internal/pkg/skuba/deployments/ssh"
	node "github.com/SUSE/skuba/pkg/skuba/actions/node/reset"

	"k8s.io/klog"
)

type resetOptions struct {
	ignorePreflightErrors string
}

// NewResetCmd creates a new cobra reset command
func NewResetCmd() *cobra.Command {
	resetOptions := resetOptions{}
	target := ssh.Target{}

	cmd := cobra.Command{
		Use:   "reset",
		Short: "Resets the node to it's state prior to running join or bootstrap",
		Run: func(cmd *cobra.Command, args []string) {
			resetConfiguration := deployments.ResetConfiguration{
				KubeadmExtraArgs: map[string]string{"ignore-preflight-errors": resetOptions.ignorePreflightErrors},
			}

			if err := node.Reset(resetConfiguration, target.GetDeployment("")); err != nil {
				klog.Fatalf("error resetting node: %s", err)
			}

			fmt.Println("successfully reset node to state prior to running bootstrap or join")

		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().AddFlagSet(target.GetFlags())
	cmd.Flags().StringVar(&resetOptions.ignorePreflightErrors, "ignore-preflight-errors", "", "Comma separated list of preflight errors to ignore")

	return &cmd
}
