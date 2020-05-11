package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPVKillerKill(t *testing.T) {
	p, err := NewPVKiller()
	assert.Nil(t, err)
	err = p.Kill()
	assert.Nil(t, err)
}
