package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewNodeKiller("test-node")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestNodeKillerKillWithBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewNodeKiller("test-node")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}
