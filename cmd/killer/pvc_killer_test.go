package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKillUnBoundPVC(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewPVCKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestPVCKillerKillAll(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewPVCKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestPVCKillerKillHalf(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewPVCKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand().SetHalf()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestPVCKillerNewPVCKiller(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewPVCKiller("default")
	assert.Nil(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "default", k.namespace)
	assert.NotNil(t, k.client)
	assert.False(t, k.dryRun)
	assert.False(t, k.mafia)
	assert.False(t, k.half)
}

func TestPVCKillerDryRun(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewPVCKiller("default")
	assert.Nil(t, err)
	result := k.DryRun()
	assert.True(t, k.dryRun)
	assert.Equal(t, k, result) // Should return self for chaining
}

func TestPVCKillerBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewPVCKiller("default")
	assert.Nil(t, err)
	result := k.BlackHand()
	assert.True(t, k.mafia)
	assert.Equal(t, k, result) // Should return self for chaining
}

func TestPVCKillerSetHalf(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewPVCKiller("default")
	assert.Nil(t, err)
	result := k.SetHalf()
	assert.True(t, k.half)
	assert.Equal(t, k, result) // Should return self for chaining
}
