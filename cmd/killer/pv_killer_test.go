package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKillUnBoundPV(t *testing.T) {
	p, err := NewPVKiller()
	assert.Nil(t, err)
	err = p.KillUnBoundPV(false)
	assert.Nil(t, err)
}
