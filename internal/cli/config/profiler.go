package config

import "k8s.io/cli-runtime/pkg/genericclioptions"

type ProfilerConfig struct {
	TargetConfig *TargetConfig
	JobConfig    *JobConfig
	ConfigFlags  *genericclioptions.ConfigFlags
}
