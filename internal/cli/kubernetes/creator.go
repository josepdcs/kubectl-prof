package kubernetes

import (
	"context"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes/job"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"os"
)

type JobCreator struct {
	clientSet kubernetes.Interface
}

// NewJobCreator returns new implementation of Creator
func NewJobCreator(clientSet kubernetes.Interface) *JobCreator {
	return &JobCreator{
		clientSet: clientSet,
	}
}

var jobType = func(language api.ProgrammingLanguage, tool api.ProfilingTool) (job.Creator, error) {
	return job.Get(language, tool)
}

func (c JobCreator) CreateProfilingJob(targetPod *v1.Pod, cfg *config.ProfilerConfig, ctx context.Context) (string, *batchv1.Job, error) {
	j, err := jobType(cfg.Target.Language, cfg.Target.ProfilingTool)
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
