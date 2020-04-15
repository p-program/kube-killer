package pkg

type KubeKiller struct {
}

func NewKubeKiller() *KubeKiller {
	killer := KubeKiller{}
	return &killer
}
