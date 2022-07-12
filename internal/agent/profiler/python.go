package profiler

import (
	"bytes"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	utils2 "github.com/josepdcs/kubectl-prof/internal/agent/utils"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

const (
	pySpyLocation = "/app/py-spy"
)

var pyResultFile = func(job *config.ProfilingJob) string {
	if stringUtils.IsBlank(job.FileName) {
		switch job.OutputType {
		case api.SpeedScope:
			return "/tmp/speedscope.json"
		case api.ThreadDump:
			return "/tmp/threaddump.txt"
		default:
			return "/tmp/flamegraph.svg"
		}
	}
	return "/tmp/" + job.FileName
}

type PythonProfiler struct{}

func NewPythonProfiler() *PythonProfiler {
	return &PythonProfiler{}
}

func (p *PythonProfiler) SetUp(job *config.ProfilingJob) error {
	return nil
}

func (p *PythonProfiler) Invoke(job *config.ProfilingJob) error {
	pid, err := utils2.ContainerPID(job, true)
	if err != nil {
		return err
	}
	utils2.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	var cmd *exec.Cmd
	var out bytes.Buffer
	var stderr bytes.Buffer

	fileName := pyResultFile(job)
	output := string(job.OutputType)
	switch job.OutputType {
	case api.FlameGraph, api.SpeedScope:
		cmd = utils2.Command(pySpyLocation, "record", "-p", pid, "-o", fileName, "-d", duration, "-s", "-t", "-f", output)
	case api.ThreadDump:
		cmd = utils2.Command(pySpyLocation, "dump", "-p", pid)
	}

	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		utils2.PublishLogEvent(api.ErrorLevel, stderr.String())
		return fmt.Errorf("could not launch profiler: %w", err)
	}

	if job.OutputType == api.ThreadDump {
		err := ioutil.WriteFile(fileName, out.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("could not save thread dump to file: %w", err)
		}
	}

	return utils2.Publish(job.Compressor, fileName, job.OutputType)
}

func (p *PythonProfiler) CleanUp(job *config.ProfilingJob) error {
	fileName := pyResultFile(job)
	err := os.Remove(fileName + api.GetExtensionFileByCompressor[job.Compressor])
	if err != nil {
		utils2.PublishLogEvent(api.WarnLevel, fmt.Sprintf("file could no be removed: %s", err))
	}
	return os.Remove(fileName)
}
