package killer

type PVCKiller struct {
}

func NewPVCKiller() *PVCKiller {
	k := PVCKiller{}
	return &k
}

func (k *PVCKiller) KillUnBoundPVC() {
	// clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	// if err != nil {
	// 	panic(err.Error())
	// }
	// clientset.
}
