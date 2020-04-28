package killer

import (
	pkg "github.com/p-program/kube-killer/pkg/condition"
	appv1 "k8s.io/api/apps/v1"
)

func init() {

}

type ResourceKiller struct {
}

func (k *ResourceKiller) KillByCondition(resource interface{}, condition pkg.Condition) {
	switch resource.(type) {
	case appv1.Deployment:
		break

	}

	// resourceType core.ResourceType, name string
}

func (k *ResourceKiller) kill() {

}
