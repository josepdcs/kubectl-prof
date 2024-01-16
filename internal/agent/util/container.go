package util

import (
	"bytes"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/containerd"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/crio"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/fake"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

const ContainerRuntimeAndContainerIdMandatoryText = "container runtime and container ID are mandatory"

// Container defines how retrieving the root filesystem and the PID of a container
type Container interface {
	RootFileSystemLocation(containerID string) (string, error)
	PID(containerID string) (string, error)
}

// NormalizeContainerID returns the container ID with suffixes removed
func NormalizeContainerID(containerID string) string {
	return regexp.MustCompile("cri-o://|containerd://").ReplaceAllString(containerID, "")
}

// runtime holds the instance of Container according the given api.ContainerRuntime
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

// ContainerFileSystem returns the root path of the container filesystem
func ContainerFileSystem(r api.ContainerRuntime, containerID string) (string, error) {
	if r == "" || containerID == "" {
		return "", errors.New(ContainerRuntimeAndContainerIdMandatoryText)
	}
	c, err := runtime(r)
	if err != nil {
		return "", err
	}
	return c.RootFileSystemLocation(containerID)
}

// ContainerPID returns the PID of the container
// Deprecated: use GetCandidatePIDs instead
func ContainerPID(job *job.ProfilingJob) (string, error) {
	if job.ContainerRuntime == "" || job.ContainerID == "" {
		return "", errors.New(ContainerRuntimeAndContainerIdMandatoryText)
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

type ChildPIDGetter interface {
	get(pid string) string
}

type childPIDGetter struct {
}

func newChildPIDGetter() ChildPIDGetter {
	return &childPIDGetter{}
}

// get returns the PID of the child process of the given PID
func (c childPIDGetter) get(pid string) string {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("pgrep", "-P", pid)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	output := strings.TrimSpace(out.String())
	if err == nil && stringUtils.IsNotBlank(output) {
		return output
	}
	return ""
}

var childPIDGetterInstance = newChildPIDGetter()

// getRealPID returns the child PID of the given PID if found,
// otherwise returns the same one.
func getRealPID(pid string) string {
	child := childPIDGetterInstance.get(pid)
	if stringUtils.IsNotBlank(child) {
		pids := strings.Split(child, "\n")
		if len(pids) > 1 {
			log.WarningLogLn(fmt.Sprintf("Detected more than one child process %v for PID: %s", pids, pid))
			return pid
		}
		return getRealPID(child)
	}
	return pid
}

// GetCandidatePIDs returns the candidate PIDs to be profiled
func GetCandidatePIDs(job *job.ProfilingJob) ([]string, error) {
	if job.ContainerRuntime == "" || job.ContainerID == "" {
		return nil, errors.New(ContainerRuntimeAndContainerIdMandatoryText)
	}
	c, err := runtime(job.ContainerRuntime)
	if err != nil {
		return nil, err
	}
	pid, err := c.PID(job.ContainerID)
	if err != nil {
		return nil, err
	}

	// When applications launch subprocesses, their PIDs need to be identified for profiling.
	var pidsToProfile []string
	fillWithChildrenPIDs(pid, &pidsToProfile)
	return pidsToProfile, nil
}

// fillWithChildrenPIDs fills the given slice with the PIDs of the children processes of the given PID
func fillWithChildrenPIDs(pid string, pidsToProfile *[]string) {
	child := childPIDGetterInstance.get(pid)
	if stringUtils.IsNotBlank(child) {
		pids := strings.Split(child, "\n")
		if len(pids) > 1 {
			log.DebugLogLn(fmt.Sprintf("Detected more than one child process %v for PID: %s", pids, pid))
			for _, p := range pids {
				fillWithChildrenPIDs(p, pidsToProfile)
			}
			return
		}
		fillWithChildrenPIDs(child, pidsToProfile)
	} else {
		*pidsToProfile = append(*pidsToProfile, pid)
	}
}

// get command line of a process
// ps 3130 | awk -v p='COMMAND' 'NR==1 {n=index($0, p); next} {print substr($0, n)}'
