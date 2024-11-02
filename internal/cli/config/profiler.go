package config

import (
	"errors"

	"github.com/josepdcs/kubectl-prof/api"
)

// ProfilerConfig encapsulates the profiler configuration.
type ProfilerConfig struct {
	Target   *TargetConfig
	Job      *JobConfig
	LogLevel api.LogLevel
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

	if p.Job == nil {
		return nil, errors.New("JobConfig is missing")
	}

	return p, nil

}

// WithJob sets the Job
func WithJob(jobConfig *JobConfig) Option {
	return func(p *ProfilerConfig) {
		p.Job = jobConfig
	}
}

// WithLogLevel sets the level
func WithLogLevel(logLevel api.LogLevel) Option {
	return func(p *ProfilerConfig) {
		p.LogLevel = logLevel
	}
}

// DeepCopy returns a deep copy of the ProfilerConfig
func (p *ProfilerConfig) DeepCopy() *ProfilerConfig {
	return &ProfilerConfig{
		Target:   p.Target.DeepCopy(),
		Job:      p.Job.DeepCopy(),
		LogLevel: p.LogLevel,
	}
}
