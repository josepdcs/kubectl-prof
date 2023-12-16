package config

import (
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	apiv1 "k8s.io/api/core/v1"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
)

type TargetConfig struct {
	Namespace            string
	PodName              string
	ContainerName        string
	ContainerID          string
	Event                api.ProfilingEvent
	Duration             time.Duration
	Interval             time.Duration
	Id                   string
	LocalPath            string
	Alpine               bool
	DryRun               bool
	Image                string
	ContainerRuntime     api.ContainerRuntime
	ContainerRuntimePath string
	Language             api.ProgrammingLanguage
	Compressor           compressor.Type
	ImagePullSecret      string
	ServiceAccountName   string
	ProfilingTool        api.ProfilingTool
	OutputType           api.OutputType
	ImagePullPolicy      apiv1.PullPolicy

	ExtraTargetOptions
}

type ExtraTargetOptions struct {
	PrintLogs                bool
	GracePeriodEnding        time.Duration
	HeapDumpSplitInChunkSize string
	PoolSizeRetrieveChunks   int
	RetrieveFileRetries      int
	PID                      string
	Pgrep                    string
}
