package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewServiceKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestServiceKillerKillAll(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewServiceKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestServiceKillerKillHalf(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewServiceKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand().SetHalf()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestServiceKillerNewServiceKiller(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewServiceKiller("default")
	assert.Nil(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "default", k.namespace)
	assert.NotNil(t, k.client)
	assert.False(t, k.dryRun)
	assert.False(t, k.mafia)
	assert.False(t, k.half)
}

func TestServiceKillerDryRun(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewServiceKiller("default")
	assert.Nil(t, err)
	result := k.DryRun()
	assert.True(t, k.dryRun)
	assert.Equal(t, k, result) // Should return self for chaining
}

func TestServiceKillerBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewServiceKiller("default")
	assert.Nil(t, err)
	result := k.BlackHand()
	assert.True(t, k.mafia)
	assert.Equal(t, k, result) // Should return self for chaining
}

func TestServiceKillerSetHalf(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewServiceKiller("default")
	assert.Nil(t, err)
	result := k.SetHalf()
	assert.True(t, k.half)
	assert.Equal(t, k, result) // Should return self for chaining
}
