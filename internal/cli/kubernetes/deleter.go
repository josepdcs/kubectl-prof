package kubernetes

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type deleter struct {
	clientSet kubernetes.Interface
}

// NewDeleter returns new implementation of Deleter
func NewDeleter(clientSet kubernetes.Interface) Deleter {
	return &deleter{
		clientSet: clientSet,
	}
}

func (d deleter) DeleteProfilingJob(job *batchv1.Job, ctx context.Context) error {
	deleteStrategy := metav1.DeletePropagationForeground
	return d.clientSet.
		BatchV1().
		Jobs(job.Namespace).
		Delete(ctx, job.Name, metav1.DeleteOptions{
			PropagationPolicy: &deleteStrategy,
		})
}
