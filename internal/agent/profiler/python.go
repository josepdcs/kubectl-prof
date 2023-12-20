package profiler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/alitto/pond"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"os/exec"
	"strconv"
	"time"
)

const (
	pySpyLocation         = "/app/py-spy"
	pySpyDelayBetweenJobs = 2 * time.Second
)

var pythonCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	output := job.OutputType
	if job.OutputType == api.FlameGraph {
		// overrides to Raw
		output = api.Raw
	}

	switch output {
	case api.FlameGraph, api.SpeedScope, api.Raw:
		interval := strconv.Itoa(int(job.Interval.Seconds()))
		args := []string{"record"}
		args = append(args, "-p", pid, "-o", fileName, "-d", interval, "-s", "-t", "-f", string(output))
		return util.Command(pySpyLocation, args...)
	// api.ThreadDump:
	default:
		args := []string{"dump"}
		args = append(args, "-p", pid)
		return util.Command(pySpyLocation, args...)
	}
}

type PythonProfiler struct {
	targetPIDs []string
	delay      time.Duration
	cmd        *exec.Cmd
	PythonManager
}

type PythonManager interface {
	invoke(job *job.ProfilingJob, pid string, fileName string) (error, string, time.Duration)
	handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string) error
	publishResult(compressor compressor.Type, fileName string, outputType api.OutputType) error
	cleanUp(cmd *exec.Cmd)
}

type pythonManager struct {
}

func NewPythonProfiler() *PythonProfiler {
	return &PythonProfiler{
		delay:         pySpyDelayBetweenJobs,
		PythonManager: &pythonManager{},
	}
}

func (p *PythonProfiler) SetUp(job *job.ProfilingJob) error {
	if stringUtils.IsNotBlank(job.PID) {
		p.targetPIDs = []string{job.PID}
		return nil
	}
	pids, err := util.GetCandidatePIDs(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PIDs to be profiled: %s", pids))
	p.targetPIDs = pids

	return nil
}

func (p *PythonProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	files := make([]string, 0, len(p.targetPIDs))

	// override output type when flamegraph: it will be generated in two steps
	// 1 - get raw format
	// 2 - convert to flamegraph with Brendan Gregg's tool: flamegraph.pl
	// the flamegraph format of py-spy will be ignored since the size of this one cannot be adjusted, and we want it
	var fileName string
	if job.OutputType == api.FlameGraph {
		fileName = common.GetResultFile(common.TmpDir(), job.Tool, api.Raw)
	} else {
		fileName = common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType)
	}

	pool := pond.New(len(p.targetPIDs), 0, pond.MinWorkers(len(p.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range p.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, f, _ := p.invoke(job, pid, fileName)
			if err == nil {
				files = append(files, f)
			}
			return err
		})
		// wait a bit between jobs for not overloading the system
		time.Sleep(p.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()
	if err != nil {
		return err, time.Since(start)
	}

	// merge all files
	file.MergeFiles(fileName, files)

	err = p.handleProfilingResult(job, flamegraph.Get(job), fileName)
	if err != nil {
		return err, time.Since(start)
	}

	return p.publishResult(job.Compressor, common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType), job.OutputType), time.Since(start)
}

func (p *pythonManager) invoke(job *job.ProfilingJob, pid string, fileName string) (error, string, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	fileName = fileName + "." + pid
	cmd := pythonCommand(job, pid, fileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), "", time.Since(start)
	}

	if job.OutputType == api.ThreadDump {
		file.Write(fileName, out.String())
	}

	return nil, fileName, time.Since(start)
}

func (p *PythonProfiler) CleanUp(*job.ProfilingJob) error {
	p.cleanUp(p.cmd)

	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}

func (p *pythonManager) handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string) error {
	if job.OutputType == api.FlameGraph {
		if file.IsEmpty(fileName) {
			return errors.New("unable to generate flamegraph: no stacks found (maybe due low cpu load)")
		}
		// convert raw format to flamegraph
		err := flameGrapher.StackSamplesToFlameGraph(fileName, common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType))
		if err != nil {
			return errors.Wrap(err, "could not convert raw format to flamegraph")
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
