package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPodKillerKill(t *testing.T) {
	k, err := NewPodKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestPodKillerKillAll(t *testing.T) {
	k, err := NewPodKiller("")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestKillAllPods(t *testing.T) {
	k, err := NewPodKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.KillAllPods()
	assert.Nil(t, err)
}
