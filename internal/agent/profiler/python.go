package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	pySpyLocation = "/app/py-spy"
)

var pythonCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	switch job.OutputType {
	case api.FlameGraph, api.SpeedScope:
		interval := strconv.Itoa(int(job.Interval.Seconds()))
		args := []string{"record", "-p", pid, "-o", fileName, "-d", interval, "-s", "-t", "-f", string(job.OutputType)}
		return util.Command(pySpyLocation, args...)
	// api.ThreadDump:
	default:
		args := []string{"dump", "-p", pid}
		return util.Command(pySpyLocation, args...)
	}
}

type PythonProfiler struct {
	targetPID string
	cmd       *exec.Cmd
	PythonManager
}

type PythonManager interface {
	handleProfilingResult(job *job.ProfilingJob, fileName string, out bytes.Buffer) error
	publishResult(compressor compressor.Type, fileName string, outputType api.EventType) error
	cleanUp(cmd *exec.Cmd)
}

type pythonManager struct {
}

func NewPythonProfiler() *PythonProfiler {
	return &PythonProfiler{
		PythonManager: &pythonManager{},
	}
}

func (p *PythonProfiler) SetUp(job *job.ProfilingJob) error {
	pid, err := util.ContainerPID(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PID to be profiled: %s", pid))
	p.targetPID = pid

	return nil
}

func (p *PythonProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()
	fileName := common.GetResultFile(common.TmpDir(), job)

	var out bytes.Buffer
	var stderr bytes.Buffer

	p.cmd = pythonCommand(job, p.targetPID, fileName)
	p.cmd.Stdout = &out
	p.cmd.Stderr = &stderr
	err := p.cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return fmt.Errorf("could not launch profiler: %w; detail: %s", err, stderr.String()), time.Since(start)
	}

	err = p.handleProfilingResult(job, fileName, out)
	if err != nil {
		return err, time.Since(start)
	}

	return p.publishResult(job.Compressor, fileName, job.OutputType), time.Since(start)
}

func (p *PythonProfiler) CleanUp(*job.ProfilingJob) error {
	p.cleanUp(p.cmd)

	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}

func (p *pythonManager) handleProfilingResult(job *job.ProfilingJob, fileName string, out bytes.Buffer) error {
	if job.OutputType == api.ThreadDump {
		err := os.WriteFile(fileName, out.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("could not save thread dump to file: %w", err)
		}
	}
	return nil
}

func (p *pythonManager) publishResult(c compressor.Type, fileName string, outputType api.EventType) error {
	return util.Publish(c, fileName, outputType)
}

func (p *pythonManager) cleanUp(cmd *exec.Cmd) {
	if cmd != nil && cmd.ProcessState == nil {
		err := cmd.Process.Kill()
		if err != nil {
			log.WarningLogLn(fmt.Sprintf("unable kill process: %s", err))
		}
	}
}
