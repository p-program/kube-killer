package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceKiller(t *testing.T) {
	k, err := NewServiceKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}
