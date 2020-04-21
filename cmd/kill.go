package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kil kubernetes's resource",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
	},
}

func kill() {
	// IntVarP(&echoTimes, "times", "t", 1, "times to echo the input")
	// var resources []string
	resources := killCmd.Flags().StringArrayP("kill", "k", nil, "kill resource")
	fmt.Printf("bilibili:")
	fmt.Println(resources)

}
