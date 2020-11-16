package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFreeze(t *testing.T) {
	//prepare
	//kubectl apply -f
	icebox, err := NewIcebox("default")
	assert.Nil(t, err)
	icebox.Freeze("", "")
}
