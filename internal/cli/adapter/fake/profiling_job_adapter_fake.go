package fake

import (
	"context"
	"errors"
	"github.com/josepdcs/kubectl-prof/internal/cli/adapter"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// ProfilingJobAdapter fakes adapter.ProfilingJobAdapter for unit tests purposes
type ProfilingJobAdapter interface {
	adapter.ProfilingJobAdapter

	WithCreateProfilingJobReturnsError() ProfilingJobAdapter
	WithGetProfilingPodReturnsError() ProfilingJobAdapter
}

// profilingJobAdapter implements ProfilingJobAdapter for unit test purposes
type profilingJobAdapter struct {
	createProfilingJobReturnsError bool
	getProfilingPodReturnsError    bool
}

// NewProfilingJobAdapter returns new instance of ProfilingJobAdapter for unit test purposes
func NewProfilingJobAdapter() ProfilingJobAdapter {
	return &profilingJobAdapter{}
}

func (p profilingJobAdapter) WithCreateProfilingJobReturnsError() ProfilingJobAdapter {
	p.createProfilingJobReturnsError = true
	return p
}

func (p profilingJobAdapter) WithGetProfilingPodReturnsError() ProfilingJobAdapter {
	p.getProfilingPodReturnsError = true
	return p
}

func (p profilingJobAdapter) CreateProfilingJob(pod *v1.Pod, config *config.ProfilerConfig, ctx context.Context) (string, *batchv1.Job, error) {
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

func (p profilingJobAdapter) GetProfilingPod(config *config.ProfilerConfig, ctx context.Context, duration time.Duration) (*v1.Pod, error) {
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

func (p profilingJobAdapter) GetProfilingContainerName() string {
	return "ProfilingContainerName"
}

func (p profilingJobAdapter) DeleteProfilingJob(job *batchv1.Job, ctx context.Context) error {
	return nil
}
