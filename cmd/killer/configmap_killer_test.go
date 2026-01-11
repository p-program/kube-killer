package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigmapKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewConfigmapKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestConfigmapKillerKillAll(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewConfigmapKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestConfigmapKillerKillHalf(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewConfigmapKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand().SetHalf()
	err = k.Kill()
	assert.Nil(t, err)
}
