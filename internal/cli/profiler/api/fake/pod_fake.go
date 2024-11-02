package fake

import (
	"context"
	"errors"

	"github.com/josepdcs/kubectl-prof/internal/cli/profiler/api"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodApi fakes api.PodApi for unit tests purposes
type PodApi interface {
	api.PodApi

	WithReturnsError() PodApi
	WithReturnsEmpty() PodApi
}

// podApi implements PodApi for unit test purposes
type podApi struct {
	returnsError bool
	returnsEmpty bool
}

// NewPodApi returns new instance of PodApi for unit test purposes
func NewPodApi() PodApi {
	return &podApi{}
}

func (p *podApi) WithReturnsError() PodApi {
	p.returnsError = true
	return p
}

func (p *podApi) WithReturnsEmpty() PodApi {
	p.returnsEmpty = true
	return p
}

func (p *podApi) GetPod(context.Context, string, string) (*v1.Pod, error) {
	if p.returnsError {
		return nil, errors.New("error getting pod")
	}
	if p.returnsEmpty {
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

func (p *podApi) GetPodsByLabelSelector(_ context.Context, _, _ string) ([]v1.Pod, error) {
	if p.returnsError {
		return nil, errors.New("error getting pods")
	}
	if p.returnsEmpty {
		return []v1.Pod{}, nil
	}
	return []v1.Pod{
		{
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
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:        "ContainerName",
						ContainerID: "ContainerID",
					},
				},
			},
		},
		{
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
				Phase: v1.PodPending,
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:        "ContainerName",
						ContainerID: "ContainerID",
					},
				},
			},
		},
	}, nil
}
