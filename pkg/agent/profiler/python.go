package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
	"io/ioutil"
	"os/exec"
	"strconv"
)

const (
	pySpyLocation = "/app/py-spy"
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
	var cmd *exec.Cmd
	var out bytes.Buffer
	var stderr bytes.Buffer
	var fileName string

	output := string(job.OutputType)
	switch job.OutputType {
	case api.FlameGraph:
		fileName = "/tmp/python.svg"
		cmd = utils.Command(pySpyLocation, "record", "-p", pid, "-o", fileName, "-d", duration, "-s", "-t", "-f", output)
	case api.SpeedScope:
		fileName = "/tmp/speedscope.json"
		cmd = utils.Command(pySpyLocation, "record", "-p", pid, "-o", fileName, "-d", duration, "-s", "-t", "-f", output)
	case api.ThreadDump:
		fileName = "/tmp/threaddump.txt"
		cmd = utils.Command(pySpyLocation, "dump", "-p", pid)
	}

	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		api.PublishLogEvent(api.ErrorLevel, stderr.String())
		return fmt.Errorf("could not launch profiler: %w", err)
	}

	if job.OutputType == api.ThreadDump {
		err := ioutil.WriteFile(fileName, out.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("could not save dump to file: %w", err)
		}
	}

	return utils.Publish(job.Compressor, fileName, job.OutputType)
}
