package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatefulSetKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewStatefulSetKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestStatefulSetKillerKillWithBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewStatefulSetKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestStatefulSetKillerKillAllStatefulSets(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewStatefulSetKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.KillAllStatefulSets()
	assert.Nil(t, err)
}

func TestStatefulSetKillerNewStatefulSetKiller(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewStatefulSetKiller("default")
	assert.Nil(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "default", k.namespace)
	assert.NotNil(t, k.client)
	assert.False(t, k.dryRun)
	assert.False(t, k.mafia)
}

func TestStatefulSetKillerDryRun(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewStatefulSetKiller("default")
	assert.Nil(t, err)
	result := k.DryRun()
	assert.True(t, k.dryRun)
	assert.Equal(t, k, result) // Should return self for chaining
}

func TestStatefulSetKillerBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewStatefulSetKiller("default")
	assert.Nil(t, err)
	result := k.BlackHand()
	assert.True(t, k.mafia)
	assert.Equal(t, k, result) // Should return self for chaining
}
