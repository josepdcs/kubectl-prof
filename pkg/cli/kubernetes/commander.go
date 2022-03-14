package kubernetes

import (
	"context"
	"fmt"
	"github.com/josepdcs/kubectl-profile/pkg/cli/config"
	"github.com/josepdcs/kubectl-profile/pkg/cli/kubernetes/job"
	"os"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func LaunchProfilingJob(targetPod *v1.Pod, cfg *config.ProfileConfig, ctx context.Context) (string, *batchv1.Job, error) {
	id, flameJob, err := job.Create(targetPod, cfg)
	if err != nil {
		return "", nil, fmt.Errorf("unable to create job: %w", err)
	}

	if cfg.Target.DryRun {
		err := printJob(flameJob)
		return "", nil, err
	}

	createJob, err := clientSet.
		BatchV1().
		Jobs(cfg.Job.Namespace).
		Create(ctx, flameJob, metav1.CreateOptions{})
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

func DeleteProfilingJob(job *batchv1.Job, targetDetails *config.TargetConfig, ctx context.Context) error {
	deleteStrategy := metav1.DeletePropagationForeground
	return clientSet.
		BatchV1().
		Jobs(targetDetails.Namespace).
		Delete(ctx, job.Name, metav1.DeleteOptions{
			PropagationPolicy: &deleteStrategy,
		})
}
