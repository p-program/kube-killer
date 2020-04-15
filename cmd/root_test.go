package cmd

import (
	"testing"

	homedir "github.com/mitchellh/go-homedir"
)

func TestA(t *testing.T) {
	home, err := homedir.Dir()
	if err != nil {
		t.FailNow()
	}
	t.Logf("%s", home)
}
