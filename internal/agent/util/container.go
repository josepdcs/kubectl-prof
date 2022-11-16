package util

import (
	"bytes"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/containerd"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/crio"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/fake"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

type Container interface {
	RootFileSystemLocation(containerID string) (string, error)
	PID(containerID string) (string, error)
}

func NormalizeContainerID(containerID string) string {
	return regexp.MustCompile("cri-o://|containerd://").ReplaceAllString(containerID, "")
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
	case api.FakeContainer:
		return fake.NewRuntimeFake(), nil
	case api.FakeContainerWithRootFileSystemLocationResultError:
		return fake.NewRuntimeFake().WithRootFileSystemLocationResultError(), nil
	case api.FakeContainerWithPIDResultError:
		return fake.NewRuntimeFake().WithPIDResultError(), nil
	default:
		return nil, errors.Errorf("unsupported container runtime: %s", r)
	}
}

func ContainerFileSystem(r api.ContainerRuntime, containerID string) (string, error) {
	if r == "" || containerID == "" {
		return "", errors.New("container runtime and container ID are mandatory")
	}
	c, err := runtime(r)
	if err != nil {
		return "", err
	}
	return c.RootFileSystemLocation(containerID)
}

func ContainerPID(job *job.ProfilingJob) (string, error) {
	if job.ContainerRuntime == "" || job.ContainerID == "" {
		return "", errors.New("container runtime and container ID are mandatory")
	}
	c, err := runtime(job.ContainerRuntime)
	if err != nil {
		return "", err
	}
	pid, err := c.PID(job.ContainerID)
	if err != nil {
		return "", err
	}

	// In some cases applications are executed through a shell script,
	// so the found PID in this point is the one of this script and, therefore,
	// is needed to guess the PID of the child process which is of the application
	return getRealPID(pid), nil
}

// getRealPID returns the child PID of the given PID if found,
// otherwise returns the same one.
func getRealPID(pid string) string {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := Command("pgrep", "-P", pid)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil && stringUtils.IsNotBlank(out.String()) {
		return getRealPID(strings.TrimSpace(out.String()))
	}
	return pid
}
