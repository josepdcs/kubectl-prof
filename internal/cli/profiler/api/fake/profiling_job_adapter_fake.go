package fake

import (
	"context"
	"errors"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/profiler/api"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProfilingJobApi fakes api.ProfilingJobApi for unit tests purposes
type ProfilingJobApi interface {
	api.ProfilingJobApi

	WithCreateProfilingJobReturnsError() ProfilingJobApi
	WithGetProfilingPodReturnsError() ProfilingJobApi
}

// profilingJobApi implements ProfilingJobApi for unit test purposes
type profilingJobApi struct {
	createProfilingJobReturnsError bool
	getProfilingPodReturnsError    bool
}

// NewProfilingJobApi returns new instance of ProfilingJobApi for unit test purposes
func NewProfilingJobApi() ProfilingJobApi {
	return &profilingJobApi{}
}

func (p *profilingJobApi) WithCreateProfilingJobReturnsError() ProfilingJobApi {
	p.createProfilingJobReturnsError = true
	return p
}

func (p *profilingJobApi) WithGetProfilingPodReturnsError() ProfilingJobApi {
	p.getProfilingPodReturnsError = true
	return p
}

func (p *profilingJobApi) CreateProfilingJob(pod *v1.Pod, config *config.ProfilerConfig, ctx context.Context) (string, *batchv1.Job, error) {
	if p.createProfilingJobReturnsError {
		return "", nil, errors.New("error creating profiling job")
	}
	return "ID", &batchv1.Job{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       batchv1.JobSpec{},
		Status:     batchv1.JobStatus{},
	}, nil
}

func (p *profilingJobApi) GetProfilingPod(config *config.ProfilerConfig, ctx context.Context, duration time.Duration) (*v1.Pod, error) {
	if p.getProfilingPodReturnsError {
		return nil, errors.New("error getting profiling pod")
	}
	return &v1.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1.PodSpec{},
		Status:     v1.PodStatus{},
	}, nil
}

func (p *profilingJobApi) GetProfilingContainerName() string {
	return "ProfilingContainerName"
}

func (p *profilingJobApi) DeleteProfilingJob(job *batchv1.Job, ctx context.Context) error {
	return nil
}
