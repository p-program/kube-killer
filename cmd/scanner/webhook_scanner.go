package scanner

import (
	"fmt"

	"github.com/rs/zerolog/log"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type WebhookScanner struct {
	client *kubernetes.Clientset
}

func NewWebhookScanner(config *rest.Config) *WebhookScanner {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}
	return &WebhookScanner{client: client}
}

func (s *WebhookScanner) Scan() ([]ScanResult, error) {
	var results []ScanResult

	// Scan ValidatingWebhookConfigurations
	validatingWebhooks, err := s.client.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(nil, metav1.ListOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list ValidatingWebhookConfigurations")
	} else {
		for _, vwc := range validatingWebhooks.Items {
			results = append(results, s.scanValidatingWebhookConfig(&vwc)...)
		}
	}

	// Scan MutatingWebhookConfigurations
	mutatingWebhooks, err := s.client.AdmissionregistrationV1().MutatingWebhookConfigurations().List(nil, metav1.ListOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list MutatingWebhookConfigurations")
	} else {
		for _, mwc := range mutatingWebhooks.Items {
			results = append(results, s.scanMutatingWebhookConfig(&mwc)...)
		}
	}

	return results, nil
}

func (s *WebhookScanner) scanValidatingWebhookConfig(vwc *admissionregistrationv1.ValidatingWebhookConfiguration) []ScanResult {
	var results []ScanResult

	for _, webhook := range vwc.Webhooks {
		// Check 1: Webhook timeout too short
		if s.hasShortTimeout(webhook.TimeoutSeconds) {
			results = append(results, ScanResult{
				Category:      "Webhook",
				Severity:      "warning",
				Resource:      "ValidatingWebhookConfiguration",
				Namespace:     "",
				Name:          fmt.Sprintf("%s/%s", vwc.Name, webhook.Name),
				Issue:         "Webhook timeout too short",
				Description:   fmt.Sprintf("Webhook %s in %s has timeoutSeconds set to 1 or less. This may cause timeouts under load.", webhook.Name, vwc.Name),
				Recommendation: "Increase timeoutSeconds to at least 10-30 seconds, or set failurePolicy: Ignore for non-critical validations",
			})
		}

		// Check 2: No cert-manager annotation
		if s.missingCertManagerAnnotation(vwc.Annotations) {
			results = append(results, ScanResult{
				Category:      "Webhook",
				Severity:      "info",
				Resource:      "ValidatingWebhookConfiguration",
				Namespace:     "",
				Name:          fmt.Sprintf("%s/%s", vwc.Name, webhook.Name),
				Issue:         "Webhook may not use cert-manager",
				Description:   fmt.Sprintf("Webhook %s in %s doesn't appear to use cert-manager for certificate management.", webhook.Name, vwc.Name),
				Recommendation: fmt.Sprintf("Consider using cert-manager to automatically manage webhook certificates: kubectl annotate validatingwebhookconfiguration %s cert-manager.io/inject-ca-from=<namespace>/<certificate>", vwc.Name),
			})
		}

		// Check 3: FailurePolicy with short timeout
		if webhook.FailurePolicy == nil || *webhook.FailurePolicy == admissionregistrationv1.Fail {
			if webhook.TimeoutSeconds != nil && *webhook.TimeoutSeconds <= 5 {
				results = append(results, ScanResult{
					Category:      "Webhook",
					Severity:      "warning",
					Resource:      "ValidatingWebhookConfiguration",
					Namespace:     "",
					Name:          fmt.Sprintf("%s/%s", vwc.Name, webhook.Name),
					Issue:         "Webhook with short timeout and Fail policy",
					Description:   fmt.Sprintf("Webhook %s has short timeout but failurePolicy: Fail. This may block API operations.", webhook.Name),
					Recommendation: "Consider setting failurePolicy: Ignore for non-critical validations, or increase timeoutSeconds",
				})
			}
		}
	}

	return results
}

func (s *WebhookScanner) scanMutatingWebhookConfig(mwc *admissionregistrationv1.MutatingWebhookConfiguration) []ScanResult {
	var results []ScanResult

	for _, webhook := range mwc.Webhooks {
		// Check 1: Webhook timeout too short
		if s.hasShortTimeout(webhook.TimeoutSeconds) {
			results = append(results, ScanResult{
				Category:      "Webhook",
				Severity:      "warning",
				Resource:      "MutatingWebhookConfiguration",
				Namespace:     "",
				Name:          fmt.Sprintf("%s/%s", mwc.Name, webhook.Name),
				Issue:         "Webhook timeout too short",
				Description:   fmt.Sprintf("Webhook %s in %s has timeoutSeconds set to 1 or less. This may cause timeouts under load.", webhook.Name, mwc.Name),
				Recommendation: "Increase timeoutSeconds to at least 10-30 seconds, or set failurePolicy: Ignore for non-critical validations",
			})
		}

		// Check 2: No cert-manager annotation
		if s.missingCertManagerAnnotation(mwc.Annotations) {
			results = append(results, ScanResult{
				Category:      "Webhook",
				Severity:      "info",
				Resource:      "MutatingWebhookConfiguration",
				Namespace:     "",
				Name:          fmt.Sprintf("%s/%s", mwc.Name, webhook.Name),
				Issue:         "Webhook may not use cert-manager",
				Description:   fmt.Sprintf("Webhook %s in %s doesn't appear to use cert-manager for certificate management.", webhook.Name, mwc.Name),
				Recommendation: fmt.Sprintf("Consider using cert-manager to automatically manage webhook certificates: kubectl annotate mutatingwebhookconfiguration %s cert-manager.io/inject-ca-from=<namespace>/<certificate>", mwc.Name),
			})
		}

		// Check 3: FailurePolicy with short timeout
		if webhook.FailurePolicy == nil || *webhook.FailurePolicy == admissionregistrationv1.Fail {
			if webhook.TimeoutSeconds != nil && *webhook.TimeoutSeconds <= 5 {
				results = append(results, ScanResult{
					Category:      "Webhook",
					Severity:      "warning",
					Resource:      "MutatingWebhookConfiguration",
					Namespace:     "",
					Name:          fmt.Sprintf("%s/%s", mwc.Name, webhook.Name),
					Issue:         "Webhook with short timeout and Fail policy",
					Description:   fmt.Sprintf("Webhook %s has short timeout but failurePolicy: Fail. This may block API operations.", webhook.Name),
					Recommendation: "Consider setting failurePolicy: Ignore for non-critical validations, or increase timeoutSeconds",
				})
			}
		}
	}

	return results
}

func (s *WebhookScanner) hasShortTimeout(timeoutSeconds *int32) bool {
	if timeoutSeconds == nil {
		return false // Default is 10 seconds, which is reasonable
	}
	return *timeoutSeconds <= 1
}

func (s *WebhookScanner) missingCertManagerAnnotation(annotations map[string]string) bool {
	if annotations == nil {
		return true
	}

	// Check for cert-manager annotations
	for key := range annotations {
		if key == "cert-manager.io/inject-ca-from" || 
		   key == "cert-manager.io/inject-ca-from-secret" {
			return false
		}
	}
	return true
}

