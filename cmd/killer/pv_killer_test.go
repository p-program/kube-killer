package killer

import "testing"

func TestKillUnBoundPV(t *testing.T) {
	p, err := NewPVKiller()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = p.KillUnBoundPV()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
