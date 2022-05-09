package profiler

import (
	"bytes"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
	"strconv"
)

const (
	rbspyLocation = "/app/rbspy"
)

var rbResultFile = func(job *config.ProfilingJob) string {
	if stringUtils.IsBlank(job.FileName) {
		return "/tmp/" + job.FileName
	}
	return "/tmp/flamegraph.svg"
}

type RubyProfiler struct{}

func (r *RubyProfiler) SetUp(job *config.ProfilingJob) error {
	return nil
}

func (r *RubyProfiler) Invoke(job *config.ProfilingJob) error {
	pid, err := utils.ContainerPID(job, true)
	if err != nil {
		return err
	}
	api.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	filName := rbResultFile(job)
	duration := strconv.Itoa(int(job.Duration.Seconds()))
	cmd := utils.Command(rbspyLocation, "record", "--pid", pid, "--file", filName, "--duration", duration, "--format", "flamegraph")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		api.PublishLogEvent(api.ErrorLevel, stderr.String())
		return fmt.Errorf("could not launch profiler: %w", err)
	}

	return utils.Publish(job.Compressor, filName, job.OutputType)
}
