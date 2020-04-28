package killer

import (
	"github.com/spf13/cobra"
)

var resources []string

func init() {
	// flags := NewKillCommand.Flags()
	// flags.StringArrayVarP(&resources, "kill", "k", nil, "kill resource")
	// NewKillCommand.MarkFlagRequired("kill")
}

func NewKillCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "kill",
		Short: "Kill kubernetes's resource",
		Long:  ``,
		// Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// kill()
		}}
}
