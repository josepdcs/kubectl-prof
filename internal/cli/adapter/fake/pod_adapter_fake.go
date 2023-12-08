package fake

import (
	"context"
	"errors"
	"github.com/josepdcs/kubectl-prof/internal/cli/adapter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodAdapter fakes adapter.PodAdapter for unit tests purposes
type PodAdapter interface {
	adapter.PodAdapter

	WithGetPodReturnsError() PodAdapter
	WithGetPodReturnsAnInvalidPod() PodAdapter
}

// podAdapter implements PodAdapter for unit test purposes
type podAdapter struct {
	getPodReturnsError        bool
	getPodReturnsAnInvalidPod bool
}

// NewPodAdapter returns new instance of PodAdapter for unit test purposes
func NewPodAdapter() PodAdapter {
	return &podAdapter{}
}

// WithGetPodReturnsError configures the method GetPod for returning an error instead of expected Pod
func (p *podAdapter) WithGetPodReturnsError() PodAdapter {
	p.getPodReturnsError = true
	return p
}

func (p *podAdapter) WithGetPodReturnsAnInvalidPod() PodAdapter {
	p.getPodReturnsAnInvalidPod = true
	return p
}

func (p *podAdapter) GetPod(string, string, context.Context) (*v1.Pod, error) {
	if p.getPodReturnsError {
		return nil, errors.New("error getting pod")
	}
	if p.getPodReturnsAnInvalidPod {
		return nil, nil
	}
	return &v1.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "ContainerName",
				},
			},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:        "ContainerName",
					ContainerID: "ContainerID",
				},
			},
		},
	}, nil
}
