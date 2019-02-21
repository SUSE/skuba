package main

import (
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd(args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use: "caaspctl",
	}

	cmd.PersistentFlags().StringP("salt-path", "s", "", "salt root path where the states folder is present")
	cmd.PersistentFlags().StringP("user", "u", "root", "user identity used to connect to target")
	cmd.PersistentFlags().Bool("sudo", false, "run remote command via sudo")

	cmd.MarkFlagRequired("salt-path")

	cmd.AddCommand(
		newInitCmd(),
		newBootstrapCmd(),
		newJoinCmd(),
	)

	return cmd
}

func main() {
	cmd := newRootCmd(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
