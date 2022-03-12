package details

import (
	"github.com/josepdcs/kubectl-profile/api"
	"time"
)

type ProfilingJob struct {
	Duration          time.Duration
	ID                string
	ContainerRuntime  api.ContainerRuntime
	ContainerID       string
	ContainerName     string
	PodUID            string
	Language          api.ProgrammingLanguage
	TargetProcessName string
	Event             api.ProfilingEvent
}
