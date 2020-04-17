package pkg

// Condition is a interface,
// you must implement this interface to compare the old “condition” to the new “condition”.
// Then you will know wether to kill the kubernetes‘s resource.
type Condition interface {
	NeedToKill() bool
}
