package cmd

import (
	"log"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/cli-runtime/pkg/genericclioptions"
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
		Short: "View CaaSP cluster status",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
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

func (o *CaaspOptions) Run() error {
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

	log.Printf("Node list: %v\n", nodeList)

	return nil
}
