package sinks

type DeploymentKiller struct {
}

func NewDeploymentKiller() *DeploymentKiller {
	killer := DeploymentKiller{}
	return &killer
}
