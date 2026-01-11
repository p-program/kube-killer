package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewJobKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestJobKillerKillAll(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewJobKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestJobKillerKillHalf(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewJobKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand().SetHalf()
	err = k.Kill()
	assert.Nil(t, err)
}
