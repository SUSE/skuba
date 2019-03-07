package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kubectlget "k8s.io/kubernetes/pkg/kubectl/cmd/get"
)

func newCmdCaaspNodes(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewCmdCaaspOptions(streams)

	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "View CaaSP cluster status",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.RunNodes(); err != nil {
				return err
			}

			return nil
		},
	}

	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func (o *CaaspOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	var err error
	o.rawConfig, err = o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	return nil
}

func (o *CaaspOptions) Validate() error {
	return nil
}

func (o *CaaspOptions) RunNodes() error {
	clientConfig, err := clientcmd.NewDefaultClientConfig(o.rawConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		log.Fatal("could not load kubeconfig file")
	}

	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		log.Fatal("could not create API client")
	}

	nodeList, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Fatal("could not retrieve node list")
	}

	outputFormat := "custom-columns=NAME:.metadata.name,OS-IMAGE:.status.nodeInfo.osImage,KERNEL-VERSION:.status.nodeInfo.kernelVersion,CONTAINER-RUNTIME:.status.nodeInfo.containerRuntimeVersion,HAS-UPDATES:.metadata.annotations.caasp\\.suse\\.com/has-updates,HAS-DISRUPTIVE-UPDATES:.metadata.annotations.caasp\\.suse\\.com/has-disruptive-updates"

	printFlags := kubectlget.NewGetPrintFlags()
	printFlags.OutputFormat = &outputFormat

	printer, err := printFlags.ToPrinter()
	if err != nil {
		log.Fatal("could not create printer")
	}
	printer.PrintObj(nodeList, os.Stdout)

	return nil
}
