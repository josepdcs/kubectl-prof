package utils

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils/containerd"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils/crio"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils/docker"
	"github.com/pkg/errors"
	"regexp"
)

func NormalizeContainerID(containerID string) string {
	return regexp.MustCompile("docker://|cri-o://|containerd://").ReplaceAllString(containerID, "")
}

func ContainerFileSystem(runtime api.ContainerRuntime, containerID string) (string, error) {
	if runtime == "" || containerID == "" {
		return "", errors.New("container runtime and container ID are mandatory")
	}
	switch runtime {
	case api.Crio:
		return crio.RootFileSystemLocation(containerID)
	case api.Containerd:
		return containerd.RootFileSystemLocation(containerID)
	case api.Docker:
		return docker.RootFileSystemLocation(containerID)
	default:
		return "", errors.Errorf("unsupported container runtime: %s", runtime)
	}
}

func ContainerPID(job *config.ProfilingJob, PPID bool) (string, error) {
	if job.ContainerRuntime == "" || job.ContainerID == "" {
		return "", errors.New("container runtime and container ID are mandatory")
	}
	switch job.ContainerRuntime {
	case api.Crio:
		return crio.PID(job.ContainerID)
	case api.Containerd:
		return containerd.PID(job.ContainerID)
	case api.Docker:
		if PPID {
			return docker.PPID(job)
		}
		return docker.PID(job)
	default:
		return "", errors.Errorf("unsupported container runtime: %s", job.ContainerRuntime)
	}
}
