package kubernetes

import (
	"context"
	"github.com/josepdcs/kubectl-profile/pkg/cli/config"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deleter struct {
}

//NewDeleter returns new implementation of Deleter
func NewDeleter() *deleter {
	return &deleter{}
}

func (d deleter) DeleteProfilingJob(job *batchv1.Job, targetDetails *config.TargetConfig, ctx context.Context) error {
	deleteStrategy := metav1.DeletePropagationForeground
	return clientSet.
		BatchV1().
		Jobs(targetDetails.Namespace).
		Delete(ctx, job.Name, metav1.DeleteOptions{
			PropagationPolicy: &deleteStrategy,
		})
}
