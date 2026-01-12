package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewDeploymentKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestDeploymentKillerKillAll(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewDeploymentKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestDeploymentKillerKillHalf(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewDeploymentKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand().SetHalf()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestDeploymentKillerKillAllDeployments(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewDeploymentKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.KillAllDeployments()
	assert.Nil(t, err)
}

func TestDeploymentKillerKillHalfDeployments(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewDeploymentKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.KillHalfDeployments()
	assert.Nil(t, err)
}

func TestDeploymentKillerNewDeploymentKiller(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewDeploymentKiller("default")
	assert.Nil(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "default", k.namespace)
	assert.NotNil(t, k.client)
	assert.False(t, k.dryRun)
	assert.False(t, k.mafia)
	assert.False(t, k.half)
}

func TestDeploymentKillerDryRun(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewDeploymentKiller("default")
	assert.Nil(t, err)
	result := k.DryRun()
	assert.True(t, k.dryRun)
	assert.Equal(t, k, result) // Should return self for chaining
}

func TestDeploymentKillerBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewDeploymentKiller("default")
	assert.Nil(t, err)
	result := k.BlackHand()
	assert.True(t, k.mafia)
	assert.Equal(t, k, result) // Should return self for chaining
}

func TestDeploymentKillerSetHalf(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewDeploymentKiller("default")
	assert.Nil(t, err)
	result := k.SetHalf()
	assert.True(t, k.half)
	assert.Equal(t, k, result) // Should return self for chaining
}
