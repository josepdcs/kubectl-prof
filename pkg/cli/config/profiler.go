package config

import (
	"github.com/josepdcs/kubectl-profile/api"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"strings"
)

type ProfilerConfig struct {
	Target       *TargetConfig
	Job          *JobConfig
	ConfigFlags  *genericclioptions.ConfigFlags
	LogLevel     api.LogLevel
	Privileged   bool
	Capabilities []apiv1.Capability
}

//NewProfilerConfig instance new ProfilerConfig with given parameters and default values for the rest
func NewProfilerConfig(Target *TargetConfig, Job *JobConfig, ConfigFlags *genericclioptions.ConfigFlags) *ProfilerConfig {
	return &ProfilerConfig{
		Target:      Target,
		Job:         Job,
		ConfigFlags: ConfigFlags,
		LogLevel:    api.InfoLevel,
		Privileged:  true,
	}
}

//WithLogLevel set log level
func (p *ProfilerConfig) WithLogLevel(logLevel api.LogLevel) *ProfilerConfig {
	p.LogLevel = logLevel
	return p
}

//WithPrivileged set privileged flag for running container in privileged mode
func (p *ProfilerConfig) WithPrivileged(privileged bool) *ProfilerConfig {
	p.Privileged = privileged
	return p
}

//WithCapabilities set the given capabilities separated by commas
// Ex. SYS_ADMIN,SYS_PTRACE
func (p *ProfilerConfig) WithCapabilities(capabilities string) *ProfilerConfig {
	if capabilities != "" {
		caps := strings.Split(capabilities, ",")
		for _, c := range caps {
			p.Capabilities = append(p.Capabilities, apiv1.Capability(c))
		}
	}
	return p
}
