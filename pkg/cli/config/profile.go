package config

import "k8s.io/cli-runtime/pkg/genericclioptions"

type ProfileConfig struct {
	Target      *TargetConfig
	Job         *JobConfig
	ConfigFlags *genericclioptions.ConfigFlags
}
