package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
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
	case api.FlameGraph, api.SpeedScope, api.Raw:
		interval := strconv.Itoa(int(job.Interval.Seconds()))
		args := []string{"record"}
		args = append(args, "-p", pid, "-o", fileName, "-d", interval, "-s", "-t", "-f", string(job.OutputType))
		return util.Command(pySpyLocation, args...)
	// api.ThreadDump:
	default:
		args := []string{"dump"}
		args = append(args, "-p", pid)
		return util.Command(pySpyLocation, args...)
	}
}

type PythonProfiler struct {
	targetPID string
	cmd       *exec.Cmd
	PythonManager
}

type PythonManager interface {
	handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string, out bytes.Buffer) error
	publishResult(compressor compressor.Type, fileName string, outputType api.OutputType) error
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
	// override output type when flamegraph: it will be generated in two steps
	// 1 - get raw format
	// 2 - convert to flamegraph with Brendan Gregg's tool: flamegraph.pl
	// the flamegraph format of py-spy will be ignored since the size of this one cannot be adjusted, and we want it
	copiedJob := *job
	var flameGrapher flamegraph.FrameGrapher
	if job.OutputType == api.FlameGraph {
		copiedJob.OutputType = api.Raw
		flameGrapher = flamegraph.Get(job)

	}
	fileName := common.GetResultFile(common.TmpDir(), &copiedJob)

	var out bytes.Buffer
	var stderr bytes.Buffer

	p.cmd = pythonCommand(&copiedJob, p.targetPID, fileName)
	p.cmd.Stdout = &out
	p.cmd.Stderr = &stderr
	err := p.cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return fmt.Errorf("could not launch profiler: %w; detail: %s", err, stderr.String()), time.Since(start)
	}

	err = p.handleProfilingResult(job, flameGrapher, fileName, out)
	if err != nil {
		return err, time.Since(start)
	}

	return p.publishResult(job.Compressor, common.GetResultFile(common.TmpDir(), job), job.OutputType), time.Since(start)
}

func (p *PythonProfiler) CleanUp(*job.ProfilingJob) error {
	p.cleanUp(p.cmd)

	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}

func (p *pythonManager) handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string, out bytes.Buffer) error {
	if job.OutputType == api.ThreadDump {
		err := os.WriteFile(fileName, out.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("could not save thread dump: %w", err)
		}
	} else if job.OutputType == api.FlameGraph {
		if file.IsEmpty(fileName) {
			return fmt.Errorf("unable to generate flamegraph: no stacks found (maybe due low cpu load)")
		}
		// convert raw format to flamegraph
		err := flameGrapher.StackSamplesToFlameGraph(fileName, common.GetResultFile(common.TmpDir(), job))
		if err != nil {
			return fmt.Errorf("could not convert raw format to flamegraph: %w", err)
		}
	}
	return nil
}

func (p *pythonManager) publishResult(c compressor.Type, fileName string, outputType api.OutputType) error {
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
