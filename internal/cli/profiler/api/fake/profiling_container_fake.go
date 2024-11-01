package fake

import (
	"context"
	"errors"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/profiler/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/result"
	v1 "k8s.io/api/core/v1"
)

// ProfilingContainerApi fakes api.ProfilingContainerApi for unit tests purposes
type ProfilingContainerApi interface {
	api.ProfilingContainerApi

	WithHandleProfilingContainerLogsReturnsError() ProfilingContainerApi
	WithGetRemoteFileReturnsError() ProfilingContainerApi
}

// profilingContainerApi implements ProfilingContainerApi for unit test purposes
type profilingContainerApi struct {
	handleProfilingContainerLogsReturnsError bool
	getRemoteFileReturnsError                bool
}

// NewProfilingContainerApi returns new instance of ProfilingContainerApi for unit test purposes
func NewProfilingContainerApi() ProfilingContainerApi {
	return &profilingContainerApi{}
}

// WithHandleProfilingContainerLogsReturnsError configures the method HandleProfilingContainerLogs for returning an error instead of expected channels
func (p *profilingContainerApi) WithHandleProfilingContainerLogsReturnsError() ProfilingContainerApi {
	p.handleProfilingContainerLogsReturnsError = true
	return p
}

func (p *profilingContainerApi) WithGetRemoteFileReturnsError() ProfilingContainerApi {
	p.getRemoteFileReturnsError = true
	return p
}

func (p *profilingContainerApi) HandleProfilingContainerLogs(*v1.Pod, string, api.EventHandler, context.Context) (chan bool, chan result.File, error) {
	if p.handleProfilingContainerLogsReturnsError {
		return nil, nil, errors.New("error handling profiling container logs")
	}
	done := make(chan bool, 1)
	done <- true
	resultFile := make(chan result.File, 1)
	resultFile <- result.File{
		FileName:  "filename",
		Timestamp: time.Now(),
	}
	return done, resultFile, nil
}

func (p *profilingContainerApi) GetRemoteFile(pod *v1.Pod, containerName string, remoteFile result.File, targetPodName string, target *config.TargetConfig) (string, error) {
	if p.getRemoteFileReturnsError {
		return "", errors.New("error getting remote file")
	}
	return "remote-file", nil
}
