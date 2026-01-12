package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCRKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRKiller("*.example.com", "")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestCRKillerKillWithBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRKiller("*.example.com", "")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestCRKillerNewCRKiller(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRKiller("*.example.com", "default")
	assert.Nil(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "default", k.namespace)
	assert.Equal(t, "*.example.com", k.groupPattern)
	assert.NotNil(t, k.dynamicClient)
	assert.NotNil(t, k.crdClient)
	assert.False(t, k.dryRun)
	assert.False(t, k.mafia)
}

func TestCRKillerNewCRKillerAllNamespaces(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRKiller("*.example.com", "")
	assert.Nil(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "", k.namespace)
	assert.Equal(t, "*.example.com", k.groupPattern)
}

func TestCRKillerDryRun(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRKiller("*.example.com", "default")
	assert.Nil(t, err)
	result := k.DryRun()
	assert.True(t, k.dryRun)
	assert.Equal(t, k, result) // Should return self for chaining
}

func TestCRKillerBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRKiller("*.example.com", "default")
	assert.Nil(t, err)
	result := k.BlackHand()
	assert.True(t, k.mafia)
	assert.Equal(t, k, result) // Should return self for chaining
}
