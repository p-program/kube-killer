package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var resources []string

func newKillCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill kubernetes's resource",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			kill()
		},
	}
	flags := cmd.Flags()
	flags.StringArrayVarP(&resources, "kill", "k", nil, "kill resource")
	// flags.string
	return cmd
}

func kill() {
	fmt.Printf("bilibili:")
	fmt.Println(resources)
	// resourceType := resources[0]
	// resourceName := resources[1]

}
