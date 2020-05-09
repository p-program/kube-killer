package killer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKillNonRunningPods(t *testing.T) {
	k := NewPodKiller(false, "")
	err := k.KillNonRunningPods()
	assert.Nil(t, err)
}

func TestKillAllPods(t *testing.T) {
	k := NewPodKiller(true, "")
	err := k.KillAllPods()
	assert.Nil(t, err)
}
