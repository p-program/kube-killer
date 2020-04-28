package killer

import (
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NAME = "kube-killer-"
)

type JobBuilder struct {
	Name string
	job  *batchv1.Job
}

func newJobBuilder() *JobBuilder {
	b := JobBuilder{
		job: &batchv1.Job{},
	}
	job := b.job
	var parallelism int32 = 1
	var completions int32 = 1
	var backoffLimit int32 = 2
	objectMeta := metav1.ObjectMeta{}
	job.Spec = batchv1.JobSpec{
		Parallelism:  &parallelism,
		Completions:  &completions,
		BackoffLimit: &backoffLimit,
		Template: v1.PodTemplateSpec{
			ObjectMeta: objectMeta,
		},
	}
	return &b
}

func (b *JobBuilder) AddNamespace(ns string) *JobBuilder {
	b.job.ObjectMeta.Namespace = ns
	return b
}

func (b *JobBuilder) RunAt(date time.Time) *JobBuilder {

	return b
}

func (b *JobBuilder) Build() *batchv1.Job {
	return b.job
}
