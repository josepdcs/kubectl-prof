package data

import (
	"time"

	"github.com/josepdcs/kubectl-flame/api"
)

type TargetDetails struct {
	Namespace          string
	PodName            string
	ContainerName      string
	ContainerId        string
	Event              api.ProfilingEvent
	Duration           time.Duration
	Id                 string
	FileName           string
	Alpine             bool
	DryRun             bool
	Image              string
	ContainerRuntime   api.ContainerRuntime
	DockerPath         string
	CrioPath           string
	ContainerdPath     string
	Language           api.ProgrammingLanguage
	Pgrep              string
	ImagePullSecret    string
	ServiceAccountName string
}
