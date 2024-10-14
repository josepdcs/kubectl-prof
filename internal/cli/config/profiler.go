package config

import (
	"errors"

	"github.com/josepdcs/kubectl-prof/api"
)

// ProfilerConfig encapsulates the profiler configuration.
// A profiler can be launched as Job or as EphemeralContainer (last one is experimental).
type ProfilerConfig struct {
	Target             *TargetConfig
	Job                *JobConfig
	EphemeralContainer *EphemeralContainerConfig
	LogLevel           api.LogLevel
}

// Option represents an option of the ProfilerConfig.
type Option func(s *ProfilerConfig)

// NewProfilerConfig instances a new ProfilerConfig. TargetConfig is always mandatory.
// It's task of the invoker to choice between launch a Job or an EphemeralContainer,
// but one of both (and only one) is mandatory, otherwise error is return.
func NewProfilerConfig(Target *TargetConfig, options ...Option) (*ProfilerConfig, error) {
	p := &ProfilerConfig{
		Target:   Target,
		LogLevel: api.InfoLevel,
	}

	for _, option := range options {
		option(p)
	}

	if p.Job == nil && p.EphemeralContainer == nil {
		return nil, errors.New("JobConfig and EphemeralContainerConfig are missing. One of both is mandatory")
	}

	if p.Job != nil && p.EphemeralContainer != nil {
		return nil, errors.New("JobConfig and EphemeralContainerConfig cannot be defined at the same time")
	}

	return p, nil

}

// WithJob sets the Job
func WithJob(jobConfig *JobConfig) Option {
	return func(p *ProfilerConfig) {
		p.Job = jobConfig
	}
}

// WithEphemeralContainer sets the EphemeralContainer
func WithEphemeralContainer(ephemeralContainerConfig *EphemeralContainerConfig) Option {
	return func(p *ProfilerConfig) {
		p.EphemeralContainer = ephemeralContainerConfig
	}
}

// WithLogLevel sets the LogLevel
func WithLogLevel(logLevel api.LogLevel) Option {
	return func(p *ProfilerConfig) {
		p.LogLevel = logLevel
	}
}
