package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-profile/api"
	"github.com/josepdcs/kubectl-profile/pkg/agent/config"
	"github.com/josepdcs/kubectl-profile/pkg/agent/utils"
	"os/exec"
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
	api.PublishLogEvent(api.InfoLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	cmd := exec.Command(pySpyLocation, "record", "-p", pid, "-o", pythonOutputFileName, "-d", duration, "-s", "-t")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return utils.PublishFlameGraph(pythonOutputFileName)
}
