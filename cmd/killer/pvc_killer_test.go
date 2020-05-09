package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKillUnBoundPVC(t *testing.T) {
	k := NewPVCKiller(true, "")
	err := k.KillUnBoundPVC()
	assert.Nil(t, err)
}
