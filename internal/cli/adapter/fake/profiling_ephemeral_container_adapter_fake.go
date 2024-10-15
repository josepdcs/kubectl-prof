package fake

import (
	"context"
	"errors"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/cli/adapter"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	v1 "k8s.io/api/core/v1"
)

// ProfilingEphemeralContainerAdapter fakes adapter.ProfilingEphemeralContainerAdapter for unit tests purposes
type ProfilingEphemeralContainerAdapter interface {
	adapter.ProfilingEphemeralContainerAdapter

	WithAddEphemeralContainerReturnsError() ProfilingEphemeralContainerAdapter
}

// profilingEphemeralContainerAdapter implements ProfilingEphemeralContainerAdapter for unit test purposes
type profilingEphemeralContainerAdapter struct {
	addEphemeralContainerReturnsError bool
}

// NewProfilingEphemeralContainerAdapter returns new instance of ProfilingEphemeralContainerAdapter for unit test purposes
func NewProfilingEphemeralContainerAdapter() ProfilingEphemeralContainerAdapter {
	return &profilingEphemeralContainerAdapter{}
}

func (p profilingEphemeralContainerAdapter) WithAddEphemeralContainerReturnsError() ProfilingEphemeralContainerAdapter {
	p.addEphemeralContainerReturnsError = true
	return p
}

func (p profilingEphemeralContainerAdapter) AddEphemeralContainer(pod *v1.Pod, config *config.ProfilerConfig, ctx context.Context, duration time.Duration) (*v1.Pod, error) {
	if p.addEphemeralContainerReturnsError {
		return nil, errors.New("error adding ephemeral container to pod")
	}

	ephemeralContainer := v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name:  "EphemeralContainerName",
			Image: "EphemeralContainerImage",
		},
		TargetContainerName: "TargetContainerName",
	}
	pod.Spec.EphemeralContainers = append(pod.Spec.EphemeralContainers, ephemeralContainer)

	return pod, nil
}

func (p profilingEphemeralContainerAdapter) GetEphemeralContainerName() string {
	return "EphemeralContainerName"
}
