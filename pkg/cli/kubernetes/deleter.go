package kubernetes

import (
	"context"
	"github.com/josepdcs/kubectl-prof/pkg/cli/config"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type deleter struct {
	clientSet kubernetes.Interface
}

//NewDeleter returns new implementation of Deleter
func NewDeleter(clientSet kubernetes.Interface) *deleter {
	return &deleter{
		clientSet: clientSet,
	}
}

func (d deleter) DeleteProfilingJob(job *batchv1.Job, targetDetails *config.TargetConfig, ctx context.Context) error {
	deleteStrategy := metav1.DeletePropagationForeground
	return d.clientSet.
		BatchV1().
		Jobs(targetDetails.Namespace).
		Delete(ctx, job.Name, metav1.DeleteOptions{
			PropagationPolicy: &deleteStrategy,
		})
}
