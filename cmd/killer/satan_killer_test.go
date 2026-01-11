package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSatanKillerNew(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewSatanKiller("default", false)
	assert.Nil(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "default", k.namespace)
	assert.False(t, k.allNamespaces)
}

func TestSatanKillerDryRun(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewSatanKiller("default", false)
	assert.Nil(t, err)
	k.DryRun()
	assert.True(t, k.dryRun)
}

func TestSatanKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewSatanKiller("default", false)
	assert.Nil(t, err)
	k.DryRun()
	// This will attempt to kill all excess resources in dry-run mode
	err = k.Kill()
	// We don't assert error here as it depends on cluster state
	// The test verifies the function can be called without panicking
}

func TestSatanKillerKillAllNamespaces(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewSatanKiller("", true)
	assert.Nil(t, err)
	assert.True(t, k.allNamespaces)
	k.DryRun()
	// This will attempt to kill all excess resources across all namespaces in dry-run mode
	err = k.Kill()
	// We don't assert error here as it depends on cluster state
	// The test verifies the function can be called without panicking
}
