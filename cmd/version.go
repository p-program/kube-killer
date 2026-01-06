package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// VERSION is set during build via ldflags
	VERSION = "dev"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kube-killer",
		Long:  `All software has versions. This is kube-killer's version`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("current version: v%s", VERSION)
		},
	}
}
