package adapter

import (
	"context"

	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodAdapter defines the methods for working with a pod
type PodAdapter interface {
	// GetPod returns the pod from its name and namespace
	GetPod(podName, namespace string, ctx context.Context) (*v1.Pod, error)
}

// podAdapter implements PodAdapter and wraps kubernetes.ConnectionInfo
type podAdapter struct {
	connectionInfo kubernetes.ConnectionInfo
}

// NewPodAdapter returns new instance of PodAdapter
func NewPodAdapter(connectionInfo kubernetes.ConnectionInfo) PodAdapter {
	return podAdapter{
		connectionInfo: connectionInfo,
	}
}

func (p podAdapter) GetPod(podName, namespace string, ctx context.Context) (*v1.Pod, error) {
	podObject, err := p.connectionInfo.ClientSet.
		CoreV1().
		Pods(namespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return podObject, nil
}
