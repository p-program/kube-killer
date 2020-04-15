package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	VERSION = "1.0.0"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of kube-killer",
	Long:  `All software has versions. This is kube-killer's version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s", VERSION)
	},
}
