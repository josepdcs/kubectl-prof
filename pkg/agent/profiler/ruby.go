package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-profile/pkg/agent/config"
	"github.com/josepdcs/kubectl-profile/pkg/agent/utils"
	"os/exec"
	"strconv"
)

const (
	rbspyLocation       = "/app/rbspy"
	rbspyOutputFileName = "/tmp/rbspy"
)

type RubyProfiler struct{}

func (r *RubyProfiler) SetUp(job *config.ProfilingJob) error {
	return nil
}

func (r *RubyProfiler) Invoke(job *config.ProfilingJob) error {
	pid, err := utils.FindRootProcessId(job)
	if err != nil {
		return fmt.Errorf("could not find root process ID: %w", err)
	}

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	cmd := exec.Command(rbspyLocation, "record", "--pid", pid, "--file", rbspyOutputFileName, "--duration", duration, "--format", "flamegraph")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not launch profiler: %w", err)
	}

	return utils.PublishFlameGraph(rbspyOutputFileName)
}
