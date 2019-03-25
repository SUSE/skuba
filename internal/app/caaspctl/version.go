package caaspctl

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   string
	BuildDate string
	Commit    string
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stderr, "caaspctl version: %s %s %s %s\n", Version, Commit, BuildDate, runtime.Version())
		},
	}
}
