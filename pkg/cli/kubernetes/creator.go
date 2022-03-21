package kubernetes

import (
	"context"
	"fmt"
	"github.com/josepdcs/kubectl-perf/pkg/cli/config"
	"github.com/josepdcs/kubectl-perf/pkg/cli/kubernetes/job"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"os"
)

type creator struct {
}

//NewCreator returns new implementation of Creator
func NewCreator() *creator {
	return &creator{}
}

func (c creator) CreateProfilingJob(targetPod *v1.Pod, cfg *config.ProfilerConfig, ctx context.Context) (string, *batchv1.Job, error) {
	id, profilingJob, err := job.Create(targetPod, cfg)
	if err != nil {
		return "", nil, fmt.Errorf("unable to create job: %w", err)
	}

	if cfg.Target.DryRun {
		err := printJob(profilingJob)
		return "", nil, err
	}

	createJob, err := clientSet.
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
