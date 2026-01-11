package killer

import (
	"context"
	"math/rand"
	"time"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type JobKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	half         bool
	namespace    string
}

func NewJobKiller(namespace string) (*JobKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := JobKiller{
		namespace: namespace,
		client:    clientset,
	}
	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *JobKiller) DryRun() *JobKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *JobKiller) BlackHand() *JobKiller {
	k.mafia = true
	return k
}

func (k *JobKiller) SetHalf() *JobKiller {
	k.half = true
	return k
}

func (k *JobKiller) Kill() error {
	if k.mafia {
		if k.half {
			return k.KillHalfJobs()
		}
		return k.KillAllJobs()
	}
	return k.KillCompletedJobs()
}

func (k *JobKiller) KillAllJobs() error {
	jobs, err := k.client.BatchV1().Jobs(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, job := range jobs.Items {
		log.Info().Msgf("Deleting job %s in namespace %s", job.Name, k.namespace)
		err = k.client.BatchV1().Jobs(k.namespace).Delete(context.TODO(), job.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *JobKiller) KillHalfJobs() error {
	jobs, err := k.client.BatchV1().Jobs(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(jobs.Items) == 0 {
		log.Info().Msg("No jobs to kill")
		return nil
	}
	
	// Randomly shuffle the jobs list
	jobList := jobs.Items
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(jobList), func(i, j int) {
		jobList[i], jobList[j] = jobList[j], jobList[i]
	})
	
	// Calculate how many jobs to kill (half, rounded down)
	jobsToKill := len(jobList) / 2
	if jobsToKill == 0 {
		jobsToKill = 1 // At least kill one job if there's only one
	}
	
	log.Info().Msgf("Killing %d out of %d jobs", jobsToKill, len(jobList))
	for i := 0; i < jobsToKill; i++ {
		job := jobList[i]
		log.Info().Msgf("Deleting job %s in namespace %s", job.Name, k.namespace)
		err = k.client.BatchV1().Jobs(k.namespace).Delete(context.TODO(), job.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *JobKiller) KillCompletedJobs() error {
	jobs, err := k.client.BatchV1().Jobs(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, job := range jobs.Items {
		if !k.DeserveDead(job) {
			continue
		}
		log.Info().Msgf("Deleting completed job %s in namespace %s", job.Name, k.namespace)
		err = k.client.BatchV1().Jobs(k.namespace).Delete(context.TODO(), job.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *JobKiller) DeserveDead(resource interface{}) bool {
	if k.mafia {
		return true
	}
	job := resource.(batchv1.Job)
	
	// Check if job is completed based on conditions
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == "True" {
			return true
		}
		if condition.Type == batchv1.JobFailed && condition.Status == "True" {
			return true
		}
	}
	
	// If no conditions are set, check completion based on spec and status
	// A job is considered complete if:
	// 1. It has completions specified and succeeded count matches
	// 2. Or it's a single-job (no completions) and has at least one succeeded pod
	if job.Spec.Completions != nil {
		if job.Status.Succeeded >= *job.Spec.Completions {
			return true
		}
	} else {
		// Single job completion (completions defaults to 1)
		if job.Status.Succeeded > 0 {
			return true
		}
	}
	
	// Check if job has failed (backoff limit exceeded)
	if job.Spec.BackoffLimit != nil && job.Status.Failed > *job.Spec.BackoffLimit {
		return true
	}
	
	return false
}

