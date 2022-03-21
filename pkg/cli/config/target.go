package config

import (
	"time"

	"github.com/josepdcs/kubectl-prof/api"
)

type TargetConfig struct {
	Namespace            string
	PodName              string
	ContainerName        string
	ContainerId          string
	Event                api.ProfilingEvent
	Duration             time.Duration
	Id                   string
	FileName             string
	Alpine               bool
	DryRun               bool
	Image                string
	ContainerRuntime     api.ContainerRuntime
	ContainerRuntimePath string
	Language             api.ProgrammingLanguage
	Pgrep                string
	ImagePullSecret      string
	ServiceAccountName   string
}
