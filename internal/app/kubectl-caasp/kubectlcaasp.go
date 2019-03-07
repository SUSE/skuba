package cmd

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd/api"

	"suse.com/caaspctl/internal/app/caaspctl"
)

type CaaspOptions struct {
	configFlags *genericclioptions.ConfigFlags

	rawConfig api.Config
	args      []string

	genericclioptions.IOStreams
}

func NewCmdCaaspOptions(streams genericclioptions.IOStreams) *CaaspOptions {
	return &CaaspOptions{
		configFlags: genericclioptions.NewConfigFlags(),
		IOStreams:   streams,
	}
}

func NewCmdCaasp(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewCmdCaaspOptions(streams)

	cmd := &cobra.Command{
		Use:   "caasp",
		Short: "Commands that allow you to handle a CaaSP cluster",
	}

	o.configFlags.AddFlags(cmd.Flags())

	cmd.AddCommand(
		newCmdCaaspNodes(streams), // FIXME (ereslibre): refactor
		caaspctl.NewClusterCmd(),
		caaspctl.NewNodeCmd(),
	)

	return cmd
}
