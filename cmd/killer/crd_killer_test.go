package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCRDKillerKill(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRDKiller("*.example.com")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}

func TestCRDKillerKillWithBlackHand(t *testing.T) {
	skipIfNoCluster(t)
	k, err := NewCRDKiller("*.example.com")
	assert.Nil(t, err)
	k.DryRun().BlackHand()
	err = k.Kill()
	assert.Nil(t, err)
}
