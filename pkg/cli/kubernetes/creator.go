package kubernetes

import (
	"context"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/cli/config"
	"github.com/josepdcs/kubectl-prof/pkg/cli/kubernetes/job"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"os"
)

type creator struct {
	clientSet kubernetes.Interface
}

//NewCreator returns new implementation of Creator
func NewCreator(clientSet kubernetes.Interface) *creator {
	return &creator{
		clientSet: clientSet,
	}
}

var jobType = func(language api.ProgrammingLanguage) (job.Creator, error) {
	return job.Get(language)
}

func (c creator) CreateProfilingJob(targetPod *v1.Pod, cfg *config.ProfilerConfig, ctx context.Context) (string, *batchv1.Job, error) {
	j, err := jobType(cfg.Target.Language)
	if err != nil {
		return "", nil, fmt.Errorf("unable to get type of job: %w", err)
	}
	id, profilingJob, err := j.Create(targetPod, cfg)
	if err != nil {
		return "", nil, fmt.Errorf("unable to create job: %w", err)
	}

	if cfg.Target.DryRun {
		err := printJob(profilingJob)
		return "", nil, err
	}
	createJob, err := c.clientSet.
		BatchV1().
		Jobs(cfg.Job.Namespace).
		Create(ctx, profilingJob, metav1.CreateOptions{})
	if err != nil {
		return "", nil, err
	}

	return id, createJob, nil
}

func printJob(job *batchv1.Job) error {
	encoder := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml: true,
	})

	return encoder.Encode(job, os.Stdout)
}
