package prepare

import (
	"github.com/spf13/cobra"
)

func init() {
}

func NewPrepareCommand() *cobra.Command {
return &cobra.Command{
	Use:   "prepare",
	Short: "init the kube-killer server",
	Long:  ``,
	// Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// kill()
	}}
}

