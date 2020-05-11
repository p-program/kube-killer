package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKillUnBoundPVC(t *testing.T) {
	k, err := NewPVCKiller("")
	assert.Nil(t, err)
	k.DryRun()
	err = k.Kill()
	assert.Nil(t, err)
}
