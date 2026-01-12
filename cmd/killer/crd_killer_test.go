package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCRDKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRDKiller("*.example.com")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestCRDKillerKillWithBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRDKiller("*.example.com")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestCRDKillerNewCRDKiller(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRDKiller("*.example.com")
	assert.Nil(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "*.example.com", k.groupPattern)
	assert.NotNil(t, k.crdClient)
	assert.False(t, k.dryRun)
	assert.False(t, k.mafia)
}

func TestCRDKillerDryRun(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRDKiller("*.example.com")
	assert.Nil(t, err)
	result := k.DryRun()
	assert.True(t, k.dryRun)
	assert.Equal(t, k, result) // Should return self for chaining
}

func TestCRDKillerBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRDKiller("*.example.com")
	assert.Nil(t, err)
	result := k.BlackHand()
	assert.True(t, k.mafia)
	assert.Equal(t, k, result) // Should return self for chaining
}
