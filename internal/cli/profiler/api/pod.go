package api

import (
	"context"

	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodApi defines the methods for working with a pod
type PodApi interface {
	// GetPod returns the pod from its name and namespace
	GetPod(ctx context.Context, podName, namespace string) (*v1.Pod, error)
	// GetPodsByLabelSelector returns the pods filtered by a label selector
	GetPodsByLabelSelector(ctx context.Context, namespace, labelSelector string) ([]v1.Pod, error)
}

// podApi implements PodApi and wraps kubernetes.ConnectionInfo
type podApi struct {
	connectionInfo kubernetes.ConnectionInfo
}

// NewPodApi returns new instance of PodApi
func NewPodApi(connectionInfo kubernetes.ConnectionInfo) PodApi {
	return &podApi{
		connectionInfo: connectionInfo,
	}
}

func (p *podApi) GetPod(ctx context.Context, podName, namespace string) (*v1.Pod, error) {
	podObject, err := p.connectionInfo.ClientSet.
		CoreV1().
		Pods(namespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return podObject, nil
}

func (p *podApi) GetPodsByLabelSelector(ctx context.Context, namespace, labelSelector string) ([]v1.Pod, error) {
	podList, err := p.connectionInfo.ClientSet.
		CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}
