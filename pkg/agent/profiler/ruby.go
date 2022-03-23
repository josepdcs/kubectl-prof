package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
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
	pid, err := utils.ContainerPID(job, true)
	if err != nil {
		return err
	}
	api.PublishLogEvent(api.InfoLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	cmd := utils.Command(rbspyLocation, "record", "--pid", pid, "--file", rbspyOutputFileName, "--duration", duration, "--format", "flamegraph")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		api.PublishLogEvent(api.ErrorLevel, stderr.String())
		return fmt.Errorf("could not launch profiler: %w", err)
	}

	return utils.PublishFlameGraph(rbspyOutputFileName)
}
