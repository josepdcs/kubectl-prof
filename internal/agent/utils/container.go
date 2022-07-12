package utils

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/utils/runtimes/containerd"
	"github.com/josepdcs/kubectl-prof/internal/agent/utils/runtimes/crio"
	"github.com/josepdcs/kubectl-prof/internal/agent/utils/runtimes/docker"
	"github.com/pkg/errors"
	"regexp"
)

type Container interface {
	RootFileSystemLocation(containerID string) (string, error)
	PID(containerID string) (string, error)
}

func NormalizeContainerID(containerID string) string {
	return regexp.MustCompile("docker://|cri-o://|containerd://").ReplaceAllString(containerID, "")
}

var runtime = func(r api.ContainerRuntime) (Container, error) {
	if r == "" {
		return nil, errors.New("container runtime is are mandatory")
	}
	switch r {
	case api.Crio:
		return crio.NewCrio(), nil
	case api.Containerd:
		return containerd.NewContainerd(), nil
	default:
		return nil, errors.Errorf("unsupported container runtime: %s", r)
	}
}

func ContainerFileSystem(r api.ContainerRuntime, containerID string) (string, error) {
	if r == "" || containerID == "" {
		return "", errors.New("container runtime and container ID are mandatory")
	}
	//TODO: to remove Docker support
	if r == api.Docker {
		return docker.NewDocker().RootFileSystemLocation(containerID)
	}
	c, err := runtime(r)
	if err != nil {
		return "", err
	}
	return c.RootFileSystemLocation(containerID)
}

func ContainerPID(job *config.ProfilingJob, PPID bool) (string, error) {
	if job.ContainerRuntime == "" || job.ContainerID == "" {
		return "", errors.New("container runtime and container ID are mandatory")
	}
	//TODO: to remove Docker support
	if job.ContainerRuntime == api.Docker {
		if PPID {
			return docker.NewDocker().PPID(job)
		}
		return docker.NewDocker().PID(job)
	}
	c, err := runtime(job.ContainerRuntime)
	if err != nil {
		return "", err
	}
	return c.PID(job.ContainerID)
}
