package config

import "k8s.io/cli-runtime/pkg/genericclioptions"

type ProfileConfig struct {
	Target      *TargetConfig
	Job         *JobConfig
	ConfigFlags *genericclioptions.ConfigFlags
}

func NewProfileConfig(Target *TargetConfig, Job *JobConfig, ConfigFlags *genericclioptions.ConfigFlags) *ProfileConfig {
	return &ProfileConfig{
		Target:      Target,
		Job:         Job,
		ConfigFlags: ConfigFlags,
	}
}
