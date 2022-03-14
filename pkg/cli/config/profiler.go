package config

import "k8s.io/cli-runtime/pkg/genericclioptions"

type ProfilerConfig struct {
	Target      *TargetConfig
	Job         *JobConfig
	ConfigFlags *genericclioptions.ConfigFlags
}

func NewProfilerConfig(Target *TargetConfig, Job *JobConfig, ConfigFlags *genericclioptions.ConfigFlags) *ProfilerConfig {
	return &ProfilerConfig{
		Target:      Target,
		Job:         Job,
		ConfigFlags: ConfigFlags,
	}
}
