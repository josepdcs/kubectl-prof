package config

import (
	"syscall"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	apiv1 "k8s.io/api/core/v1"
)

type TargetConfig struct {
	Namespace            string
	PodName              string
	ContainerName        string
	LabelSelector        string
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
	PoolSizeLaunchProfilingJobs int
	PrintLogs                   bool
	GracePeriodEnding           time.Duration
	HeapDumpSplitInChunkSize    string
	PoolSizeRetrieveChunks      int
	RetrieveFileRetries         int
	PID                         string
	Pgrep                       string
	NodeHeapSnapshotSignal      syscall.Signal
}

// DeepCopy returns a deep copy of the target config
func (t *TargetConfig) DeepCopy() *TargetConfig {
	return &TargetConfig{
		Namespace:            t.Namespace,
		PodName:              t.PodName,
		ContainerName:        t.ContainerName,
		LabelSelector:        t.LabelSelector,
		ContainerID:          t.ContainerID,
		Event:                t.Event,
		Duration:             t.Duration,
		Interval:             t.Interval,
		Id:                   t.Id,
		LocalPath:            t.LocalPath,
		Alpine:               t.Alpine,
		DryRun:               t.DryRun,
		Image:                t.Image,
		ContainerRuntime:     t.ContainerRuntime,
		ContainerRuntimePath: t.ContainerRuntimePath,
		Language:             t.Language,
		Compressor:           t.Compressor,
		ImagePullSecret:      t.ImagePullSecret,
		ServiceAccountName:   t.ServiceAccountName,
		ProfilingTool:        t.ProfilingTool,
		OutputType:           t.OutputType,
		ImagePullPolicy:      t.ImagePullPolicy,
		ExtraTargetOptions:   t.ExtraTargetOptions,
	}
}
