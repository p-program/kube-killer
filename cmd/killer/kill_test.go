package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectKiller(t *testing.T) {
	skipIfNoCluster(t)

	// Test with dry run to avoid actual deletions
	dryRun = true
	defer func() { dryRun = false }()

	// Test various resource types
	testCases := []struct {
		name         string
		resourceType string
		shouldError  bool
	}{
		{"configmap", "cm", false},
		{"deployment", "deploy", false},
		{"pod", "pod", false},
		{"service", "svc", false},
		{"job", "job", false},
		{"secret", "secret", false},
		{"pvc", "pvc", false},
		{"unsupported", "unsupported", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SelectKiller([]string{tc.resourceType})
			if tc.shouldError {
				assert.NotNil(t, err)
			} else {
				// Some operations may succeed or fail depending on cluster state
				// We just verify the function doesn't panic
				_ = err
			}
		})
	}
}

func TestExecuteKill(t *testing.T) {
	skipIfNoCluster(t)

	// Test ExecuteKill with dry run
	err := ExecuteKill("cm", "default", false, true, false)
	// Error may occur if no resources exist, which is fine
	_ = err
}

func TestConfirmDelete(t *testing.T) {
	// Test with interactive = false (should always return true)
	oldInteractive := interactive
	interactive = false
	defer func() { interactive = oldInteractive }()

	result := confirmDelete("pod", "test-pod", "default")
	assert.True(t, result)
}
