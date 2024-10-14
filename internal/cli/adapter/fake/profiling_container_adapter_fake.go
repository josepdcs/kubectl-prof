package fake

import (
	"context"
	"errors"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/cli/adapter"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/result"
	v1 "k8s.io/api/core/v1"
)

// ProfilingContainerAdapter fakes adapter.ProfilingContainerAdapter for unit tests purposes
type ProfilingContainerAdapter interface {
	adapter.ProfilingContainerAdapter

	WithHandleProfilingContainerLogsReturnsError() ProfilingContainerAdapter
	WithGetRemoteFileReturnsError() ProfilingContainerAdapter
}

// profilingContainerAdapter implements ProfilingContainerAdapter for unit test purposes
type profilingContainerAdapter struct {
	handleProfilingContainerLogsReturnsError bool
	getRemoteFileReturnsError                bool
}

// NewProfilingContainerAdapter returns new instance of ProfilingContainerAdapter for unit test purposes
func NewProfilingContainerAdapter() ProfilingContainerAdapter {
	return &profilingContainerAdapter{}
}

// WithHandleProfilingContainerLogsReturnsError configures the method HandleProfilingContainerLogs for returning an error instead of expected channels
func (p *profilingContainerAdapter) WithHandleProfilingContainerLogsReturnsError() ProfilingContainerAdapter {
	p.handleProfilingContainerLogsReturnsError = true
	return p
}

func (p *profilingContainerAdapter) WithGetRemoteFileReturnsError() ProfilingContainerAdapter {
	p.getRemoteFileReturnsError = true
	return p
}

func (p *profilingContainerAdapter) HandleProfilingContainerLogs(*v1.Pod, string, adapter.EventHandler, context.Context) (chan bool, chan result.File, error) {
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

func (p *profilingContainerAdapter) GetRemoteFile(pod *v1.Pod, containerName string, remoteFile result.File, target *config.TargetConfig) (string, error) {
	if p.getRemoteFileReturnsError {
		return "", errors.New("error getting remote file")
	}
	return "remote-file", nil
}
