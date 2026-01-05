package server

import (
	"github.com/spf13/cobra"
)

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run kube-killer as a Kubernetes Operator",
		Long:  `Run kube-killer as a Kubernetes Operator that watches KubeKiller CRDs and manages resource cleanup.`,
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Start the kube-killer operator",
		Long:  `Start the kube-killer operator server. This will watch for KubeKiller CRDs and execute cleanup operations based on the configured mode.`,
		Run: func(cmd *cobra.Command, args []string) {
			server := NewKubeKillerServer()
			server.Run()
		},
	}

	cmd.AddCommand(runCmd)
	return cmd
}

