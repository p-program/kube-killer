package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPVKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	p, err := NewPVKiller()
	assert.Nil(t, err)
	p.DryRun()
	err = p.Kill()
	assert.Nil(t, err)
}

func TestPVKillerKillAll(t *testing.T) {
	skipIfNoCluster(t)
	p, err := NewPVKiller()
	assert.Nil(t, err)
	p.DryRun().BlackHand()
	err = p.Kill()
	assert.Nil(t, err)
}

func TestPVKillerKillHalf(t *testing.T) {
	skipIfNoCluster(t)
	p, err := NewPVKiller()
	assert.Nil(t, err)
	p.DryRun().BlackHand().SetHalf()
	err = p.Kill()
	assert.Nil(t, err)
}

func TestPVKillerNewPVKiller(t *testing.T) {
	skipIfNoCluster(t)
	p, err := NewPVKiller()
	assert.Nil(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.client)
	assert.False(t, p.dryRun)
	assert.False(t, p.mafia)
	assert.False(t, p.half)
}

func TestPVKillerDryRun(t *testing.T) {
	skipIfNoCluster(t)
	p, err := NewPVKiller()
	assert.Nil(t, err)
	result := p.DryRun()
	assert.True(t, p.dryRun)
	assert.Equal(t, p, result) // Should return self for chaining
}

func TestPVKillerBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	p, err := NewPVKiller()
	assert.Nil(t, err)
	result := p.BlackHand()
	assert.True(t, p.mafia)
	assert.Equal(t, p, result) // Should return self for chaining
}

func TestPVKillerSetHalf(t *testing.T) {
	skipIfNoCluster(t)
	p, err := NewPVKiller()
	assert.Nil(t, err)
	result := p.SetHalf()
	assert.True(t, p.half)
	assert.Equal(t, p, result) // Should return self for chaining
}
