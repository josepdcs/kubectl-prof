package profiler

import (
	"bytes"
	"github.com/josepdcs/kubectl-profiling/internal/agent/details"
	utils2 "github.com/josepdcs/kubectl-profiling/internal/agent/utils"
	"os/exec"
	"strconv"
)

const (
	pySpyLocation        = "/app/py-spy"
	pythonOutputFileName = "/tmp/python.svg"
)

type PythonProfiler struct{}

func (p *PythonProfiler) SetUp(job *details.ProfilingJob) error {
	return nil
}

func (p *PythonProfiler) Invoke(job *details.ProfilingJob) error {
	pid, err := utils2.FindRootProcessId(job)
	if err != nil {
		return err
	}

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

	return utils2.PublishFlameGraph(pythonOutputFileName)
}
