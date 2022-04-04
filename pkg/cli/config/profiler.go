package config

import (
	"github.com/josepdcs/kubectl-prof/api"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ProfilerConfig struct {
	Target      *TargetConfig
	Job         *JobConfig
	ConfigFlags *genericclioptions.ConfigFlags
	Privileged  bool
	LogLevel    api.LogLevel
}

//NewProfilerConfig instance new ProfilerConfig with given parameters and default values for the rest
func NewProfilerConfig(Target *TargetConfig, Job *JobConfig, ConfigFlags *genericclioptions.ConfigFlags) *ProfilerConfig {
	return &ProfilerConfig{
		Target:      Target,
		Job:         Job,
		ConfigFlags: ConfigFlags,
		Privileged:  false,
		LogLevel:    api.InfoLevel,
	}
}

//WithPrivileged set privileged flag for running container in privileged mode
func (p *ProfilerConfig) WithPrivileged(privileged bool) *ProfilerConfig {
	p.Privileged = privileged
	return p
}

//WithLogLevel set log level
func (p *ProfilerConfig) WithLogLevel(logLevel api.LogLevel) *ProfilerConfig {
	p.LogLevel = logLevel
	return p
}
