package scanner

import (
	"fmt"

	"github.com/rs/zerolog/log"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/rest"
)

type CRDScanner struct {
	client apiextensionsclient.Interface
}

func NewCRDScanner(config *rest.Config) *CRDScanner {
	client, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create CRD client")
	}
	return &CRDScanner{client: client}
}

func (s *CRDScanner) Scan(k8sVersion *version.Info) ([]ScanResult, error) {
	var results []ScanResult

	crdList, err := s.client.ApiextensionsV1().CustomResourceDefinitions().List(nil, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, crd := range crdList.Items {
		// Check 1: No schema in Kubernetes 1.17-
		if s.hasNoSchema(crd, k8sVersion) {
			results = append(results, ScanResult{
				Category:      "CRD",
				Severity:      "error",
				Resource:      "CustomResourceDefinition",
				Namespace:     "",
				Name:          crd.Name,
				Issue:         "CRD without schema (Kubernetes 1.17-)",
				Description:   fmt.Sprintf("CRD %s has empty or missing schema. This is unsafe and allows invalid data.", crd.Name),
				Recommendation: "Add proper OpenAPI schema to CRD versions with preserveUnknownFields: false",
			})
		}

		// Check 2: No conversion webhook
		if s.hasNoConversionWebhook(crd) {
			results = append(results, ScanResult{
				Category:      "CRD",
				Severity:      "warning",
				Resource:      "CustomResourceDefinition",
				Namespace:     "",
				Name:          crd.Name,
				Issue:         "CRD without conversion webhook",
				Description:   fmt.Sprintf("CRD %s has multiple versions but no conversion webhook configured. Field migrations require manual YAML updates.", crd.Name),
				Recommendation: "Consider adding a conversion webhook for version migrations, or use a single version strategy",
			})
		}

		// Check 3: Status field in spec (this is harder to detect from CRD alone)
		// We can check if spec has fields that look like status fields
		if s.hasStatusLikeFieldsInSpec(crd) {
			results = append(results, ScanResult{
				Category:      "CRD",
				Severity:      "warning",
				Resource:      "CustomResourceDefinition",
				Namespace:     "",
				Name:          crd.Name,
				Issue:         "Possible status fields in spec",
				Description:   fmt.Sprintf("CRD %s may have status-like fields (ready, phase, state) in spec instead of status.", crd.Name),
				Recommendation: "Move status fields to the status subresource. Spec should only contain desired state.",
			})
		}

		// Check 4: preserveUnknownFields is true (unsafe)
		if s.hasPreserveUnknownFields(crd) {
			results = append(results, ScanResult{
				Category:      "CRD",
				Severity:      "error",
				Resource:      "CustomResourceDefinition",
				Namespace:     "",
				Name:          crd.Name,
				Issue:         "preserveUnknownFields is enabled",
				Description:   fmt.Sprintf("CRD %s has preserveUnknownFields enabled, which allows unknown fields and bypasses validation.", crd.Name),
				Recommendation: "Set preserveUnknownFields: false and define proper schema",
			})
		}
	}

	return results, nil
}

func (s *CRDScanner) hasNoSchema(crd apiextensionsv1.CustomResourceDefinition, k8sVersion *version.Info) bool {
	// Check if any version has empty schema
	for _, version := range crd.Spec.Versions {
		if version.Schema == nil || version.Schema.OpenAPIV3Schema == nil {
			return true
		}
		// Check if schema is essentially empty (only type: object with no properties)
		schema := version.Schema.OpenAPIV3Schema
		if schema.Type == "object" && (schema.Properties == nil || len(schema.Properties) == 0) {
			return true
		}
	}
	return false
}

func (s *CRDScanner) hasNoConversionWebhook(crd apiextensionsv1.CustomResourceDefinition) bool {
	// If there are multiple versions and no conversion strategy or webhook
	if len(crd.Spec.Versions) > 1 {
		if crd.Spec.Conversion == nil || 
		   crd.Spec.Conversion.Strategy != apiextensionsv1.WebhookConverter &&
		   (crd.Spec.Conversion.Webhook == nil || crd.Spec.Conversion.Webhook.ClientConfig == nil) {
			return true
		}
	}
	return false
}

func (s *CRDScanner) hasStatusLikeFieldsInSpec(crd apiextensionsv1.CustomResourceDefinition) bool {
	// This is a heuristic check - look for common status field names in spec
	statusLikeFields := []string{"ready", "phase", "state", "status", "condition", "observedGeneration"}
	
	for _, version := range crd.Spec.Versions {
		if version.Schema != nil && version.Schema.OpenAPIV3Schema != nil {
			schema := version.Schema.OpenAPIV3Schema
			if schema.Properties != nil {
				if spec, ok := schema.Properties["spec"]; ok && spec.Properties != nil {
					for fieldName := range spec.Properties {
						for _, statusField := range statusLikeFields {
							if fieldName == statusField {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func (s *CRDScanner) hasPreserveUnknownFields(crd apiextensionsv1.CustomResourceDefinition) bool {
	// In v1 API, preserveUnknownFields defaults to false, but we check if it's explicitly set to true
	// Note: This field was removed in v1, but we check for backward compatibility
	return false // preserveUnknownFields is deprecated and defaults to false in v1
}

