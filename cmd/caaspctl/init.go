package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "initialize caaspctl structure for cluster deployment",
		Run: func(*cobra.Command, []string) {
			fmt.Println("Not yet implemented")
		},
	}
}
