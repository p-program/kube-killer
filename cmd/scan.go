package cmd

import (
	"github.com/p-program/kube-killer/cmd/scanner"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	scanNamespace     string
	scanAllNamespaces bool
	scanOutput        string
)

func NewScanCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan",
		Short: "Scan Kubernetes cluster for anti-patterns and issues",
		Long: `Scan the Kubernetes cluster for common anti-patterns and issues based on 
Cloud Native Development Best Practices. This command checks for:
- CRD schema issues
- Conversion webhook problems
- Controller reconciliation loops
- Webhook configuration issues
- Owner reference problems
- And more...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ns := scanNamespace
			if scanAllNamespaces {
				ns = ""
			}

			scanner := scanner.NewClusterScanner()
			results, err := scanner.Scan(ns, scanAllNamespaces)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan cluster")
				return err
			}

			return scanner.Report(results, scanOutput)
		},
	}

	c.PersistentFlags().StringVarP(&scanNamespace, "namespace", "n", "", "Scan specific namespace (default: all)")
	c.PersistentFlags().BoolVarP(&scanAllNamespaces, "all-namespaces", "A", false, "Scan all namespaces")
	c.PersistentFlags().StringVarP(&scanOutput, "output", "o", "table", "Output format: table, json, yaml")

	return c
}
