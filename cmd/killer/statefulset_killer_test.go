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
