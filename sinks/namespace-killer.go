package sinks

type NamespaceKiller struct {
}

func NewNamespaceKiller() *NamespaceKiller {
	killer := NamespaceKiller{}
	return &killer
}
