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
	pySpyLocation        = "/app/py-spy"
	pythonOutputFileName = "/tmp/python.svg"
)

type PythonProfiler struct{}

func (p *PythonProfiler) SetUp(job *config.ProfilingJob) error {
	return nil
}

func (p *PythonProfiler) Invoke(job *config.ProfilingJob) error {
	pid, err := utils.ContainerPID(job, true)
	if err != nil {
		return err
	}
	api.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	cmd := utils.Command(pySpyLocation, "record", "-p", pid, "-o", pythonOutputFileName, "-d", duration, "-s", "-t")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		api.PublishLogEvent(api.ErrorLevel, stderr.String())
		return fmt.Errorf("could not launch profiler: %w", err)
	}

	return utils.Publish(job.Compressor, pythonOutputFileName, job.OutputType)
}
