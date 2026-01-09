package scanner

import (
	"fmt"
	"strings"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
)

type ScanResult struct {
	Category    string
	Severity    string // "error", "warning", "info"
	Resource    string
	Namespace   string
	Name        string
	Issue       string
	Description string
	Recommendation string
}

type ClusterScanner struct {
	client        *kubernetes.Clientset
	k8sVersion    *version.Info
	crdScanner    *CRDScanner
	webhookScanner *WebhookScanner
	controllerScanner *ControllerScanner
	ownerRefScanner *OwnerRefScanner
}

func NewClusterScanner() *ClusterScanner {
	config := core.GLOBAL_KUBERNETES_CONFIG
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Get Kubernetes version
	k8sVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get Kubernetes version")
	}

	scanner := &ClusterScanner{
		client:        client,
		k8sVersion:    k8sVersion,
		crdScanner:    NewCRDScanner(config),
		webhookScanner: NewWebhookScanner(config),
		controllerScanner: NewControllerScanner(config),
		ownerRefScanner: NewOwnerRefScanner(config),
	}

	return scanner
}

func (s *ClusterScanner) Scan(namespace string, allNamespaces bool) ([]ScanResult, error) {
	var results []ScanResult

	log.Info().Msg("Starting cluster scan...")
	log.Info().Str("kubernetes-version", s.k8sVersion.GitVersion).Msg("Detected Kubernetes version")

	// Scan CRDs
	log.Info().Msg("Scanning CustomResourceDefinitions...")
	crdResults, err := s.crdScanner.Scan(s.k8sVersion)
	if err != nil {
		log.Warn().Err(err).Msg("CRD scan failed")
	} else {
		results = append(results, crdResults...)
	}

	// Scan Webhooks
	log.Info().Msg("Scanning Webhooks...")
	webhookResults, err := s.webhookScanner.Scan()
	if err != nil {
		log.Warn().Err(err).Msg("Webhook scan failed")
	} else {
		results = append(results, webhookResults...)
	}

	// Scan Controllers (this would require scanning operator code, which is harder)
	// For now, we'll scan for common patterns in deployed resources
	log.Info().Msg("Scanning Controllers and Operators...")
	controllerResults, err := s.controllerScanner.Scan(namespace, allNamespaces)
	if err != nil {
		log.Warn().Err(err).Msg("Controller scan failed")
	} else {
		results = append(results, controllerResults...)
	}

	// Scan Owner References
	log.Info().Msg("Scanning Owner References...")
	ownerRefResults, err := s.ownerRefScanner.Scan(namespace, allNamespaces)
	if err != nil {
		log.Warn().Err(err).Msg("Owner reference scan failed")
	} else {
		results = append(results, ownerRefResults...)
	}

	log.Info().Int("total-issues", len(results)).Msg("Scan completed")
	return results, nil
}

func (s *ClusterScanner) Report(results []ScanResult, outputFormat string) error {
	switch strings.ToLower(outputFormat) {
	case "json":
		return s.reportJSON(results)
	case "yaml":
		return s.reportYAML(results)
	case "table":
		fallthrough
	default:
		return s.reportTable(results)
	}
}

func (s *ClusterScanner) reportTable(results []ScanResult) error {
	if len(results) == 0 {
		fmt.Println("\n✅ No issues found! Your cluster looks good.")
		return nil
	}

	fmt.Println("\n" + strings.Repeat("=", 100))
	fmt.Println("KUBE-KILLER SCAN RESULTS")
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("Total Issues Found: %d\n\n", len(results))

	// Group by category
	categories := make(map[string][]ScanResult)
	for _, result := range results {
		categories[result.Category] = append(categories[result.Category], result)
	}

	for category, categoryResults := range categories {
		fmt.Printf("\n📁 Category: %s (%d issues)\n", category, len(categoryResults))
		fmt.Println(strings.Repeat("-", 100))

		for i, result := range categoryResults {
			severityIcon := "⚠️"
			if result.Severity == "error" {
				severityIcon = "❌"
			} else if result.Severity == "info" {
				severityIcon = "ℹ️"
			}

			fmt.Printf("\n[%d] %s %s\n", i+1, severityIcon, result.Issue)
			if result.Namespace != "" {
				fmt.Printf("   Resource: %s/%s/%s\n", result.Resource, result.Namespace, result.Name)
			} else {
				fmt.Printf("   Resource: %s/%s\n", result.Resource, result.Name)
			}
			fmt.Printf("   Description: %s\n", result.Description)
			if result.Recommendation != "" {
				fmt.Printf("   💡 Recommendation: %s\n", result.Recommendation)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 100))
	return nil
}

func (s *ClusterScanner) reportJSON(results []ScanResult) error {
	// Simple JSON output
	fmt.Println("[")
	for i, result := range results {
		fmt.Printf("  {\n")
		fmt.Printf("    \"category\": \"%s\",\n", result.Category)
		fmt.Printf("    \"severity\": \"%s\",\n", result.Severity)
		fmt.Printf("    \"resource\": \"%s\",\n", result.Resource)
		fmt.Printf("    \"namespace\": \"%s\",\n", result.Namespace)
		fmt.Printf("    \"name\": \"%s\",\n", result.Name)
		fmt.Printf("    \"issue\": \"%s\",\n", result.Issue)
		fmt.Printf("    \"description\": \"%s\",\n", result.Description)
		fmt.Printf("    \"recommendation\": \"%s\"\n", result.Recommendation)
		if i < len(results)-1 {
			fmt.Printf("  },\n")
		} else {
			fmt.Printf("  }\n")
		}
	}
	fmt.Println("]")
	return nil
}

func (s *ClusterScanner) reportYAML(results []ScanResult) error {
	fmt.Println("results:")
	for _, result := range results {
		fmt.Printf("- category: %s\n", result.Category)
		fmt.Printf("  severity: %s\n", result.Severity)
		fmt.Printf("  resource: %s\n", result.Resource)
		fmt.Printf("  namespace: %s\n", result.Namespace)
		fmt.Printf("  name: %s\n", result.Name)
		fmt.Printf("  issue: %s\n", result.Issue)
		fmt.Printf("  description: %s\n", result.Description)
		fmt.Printf("  recommendation: %s\n", result.Recommendation)
		fmt.Println()
	}
	return nil
}

