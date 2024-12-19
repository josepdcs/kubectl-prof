package util

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/containerd"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/crio"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/fake"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
)

const (
	psCommand                                   = "/app/get-ps-command.sh"
	ContainerRuntimeAndContainerIdMandatoryText = "container runtime and container ID are mandatory"
)

// Container defines how retrieving the root filesystem and the PID of a container
type Container interface {
	RootFileSystemLocation(containerID string, containerRuntimePath string) (string, error)
	PID(containerID string, containerRuntimePath string) (string, error)
	CWD(containerID string, containerRuntimePath string) (string, error)
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

var commander = exec.NewCommander()

// ContainerFileSystem returns the root path of the container filesystem
func ContainerFileSystem(r api.ContainerRuntime, containerID string, containerRuntimePath string) (string, error) {
	if r == "" || containerID == "" {
		return "", errors.New(ContainerRuntimeAndContainerIdMandatoryText)
	}
	c, err := runtime(r)
	if err != nil {
		return "", err
	}
	return c.RootFileSystemLocation(containerID, containerRuntimePath)
}

// GetRootPID returns the PID of the container's root process
func GetRootPID(job *job.ProfilingJob) (string, error) {
	if job.ContainerRuntime == "" || job.ContainerID == "" {
		return "", errors.New(ContainerRuntimeAndContainerIdMandatoryText)
	}
	c, err := runtime(job.ContainerRuntime)
	if err != nil {
		return "", err
	}
	pid, err := c.PID(job.ContainerID, job.ContainerRuntimePath)
	if err != nil {
		return "", err
	}
	return pid, nil
}

// GetCWD returns the current working directory of the container's root process
func GetCWD(job *job.ProfilingJob) (string, error) {
	if job.ContainerRuntime == "" || job.ContainerID == "" {
		return "", errors.New(ContainerRuntimeAndContainerIdMandatoryText)
	}
	c, err := runtime(job.ContainerRuntime)
	if err != nil {
		return "", err
	}
	cwd, err := c.CWD(job.ContainerID, job.ContainerRuntimePath)
	if err != nil {
		return "", err
	}
	return cwd, nil
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
	cmd := commander.Command("pgrep", "-P", pid)
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

// GetCandidatePIDs returns the candidate PIDs to be profiled
func GetCandidatePIDs(job *job.ProfilingJob) ([]string, error) {
	if job.ContainerRuntime == "" || job.ContainerID == "" {
		return nil, errors.New(ContainerRuntimeAndContainerIdMandatoryText)
	}
	c, err := runtime(job.ContainerRuntime)
	if err != nil {
		return nil, err
	}
	pid, err := c.PID(job.ContainerID, job.ContainerRuntimePath)
	if err != nil {
		return nil, err
	}

	// When applications launch subprocesses, their PIDs need to be identified for profiling.
	var pidsToProfile []string
	fillWithChildrenPIDs(pid, &pidsToProfile)
	if len(pidsToProfile) == 0 {
		return nil, errors.Errorf("no PIDs found for container ID: %s", job.ContainerID)
	}
	err = filterPIDsToProfile(&pidsToProfile, job.Pgrep)
	if err != nil {
		return nil, err
	}

	// If more than one PID is detected, a notice is shown
	if len(pidsToProfile) > 1 {
		_ = log.EventLn(api.Notice,
			&api.NoticeData{
				Time: time.Now(),
				Msg: fmt.Sprintf("Detected more than one PID to profile: %v. "+
					"It will be attempt to profile all of them. "+
					"Use the --pid flag specifying the corresponding PID if you only want to profile one of them.", pidsToProfile),
			},
		)
	}

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

func filterPIDsToProfile(pids *[]string, pgrep string) error {
	if stringUtils.IsBlank(pgrep) {
		return nil
	}
	var out bytes.Buffer
	var stderr bytes.Buffer

	var filteredPIDs []string
	for _, pid := range *pids {
		cmd := commander.Command(psCommand, pid)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		output := strings.TrimSpace(out.String())
		if err != nil {
			return errors.Wrapf(err, "ps command failed with error: %s", stderr.String())
		}
		log.DebugLogLn(fmt.Sprintf("ps command output: %s", output))
		if stringUtils.ContainsIgnoreCase(output, pgrep) {
			filteredPIDs = append(filteredPIDs, pid)
		}
		out.Reset()
	}
	*pids = filteredPIDs
	return nil
}

func GetFirstCandidatePID(rootPID string) string {
	child := childPIDGetterInstance.get(rootPID)
	if stringUtils.IsBlank(child) {
		return rootPID
	}
	pids := strings.Split(child, "\n")
	if len(pids) == 1 {
		return strings.TrimSpace(pids[0])
	}
	return rootPID
}
