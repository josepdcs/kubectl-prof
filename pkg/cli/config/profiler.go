package config

import (
	"github.com/josepdcs/kubectl-prof/api"
)

type ProfilerConfig struct {
	Target   *TargetConfig
	Job      *JobConfig
	LogLevel api.LogLevel
}

//NewProfilerConfig instance new ProfilerConfig with given parameters and default values for the rest
func NewProfilerConfig(Target *TargetConfig, Job *JobConfig) *ProfilerConfig {
	return &ProfilerConfig{
		Target:   Target,
		Job:      Job,
		LogLevel: api.InfoLevel,
	}
}

//WithLogLevel set log level
func (p *ProfilerConfig) WithLogLevel(logLevel api.LogLevel) *ProfilerConfig {
	p.LogLevel = logLevel
	return p
}
