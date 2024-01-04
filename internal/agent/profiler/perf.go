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
	"github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"time"
)

const (
	perfLocation                    = "/app/perf"
	perfRecordOutputFileName        = "/tmp/perf-%s.data"
	flameGraphStackCollapseLocation = "/app/FlameGraph/stackcollapse-perf.pl"
	perfScriptOutputFileName        = "/tmp/perf-%s.out"
	perfDelayBetweenJobs            = 2 * time.Second
)

type PerfProfiler struct {
	targetPIDs []string
	delay      time.Duration
	PerfManager
}

type PerfManager interface {
	invoke(job *job.ProfilingJob, pid string) (error, string, time.Duration)
	runPerfRecord(job *job.ProfilingJob, pid string) error
	runPerfScript(job *job.ProfilingJob, pid string) error
	foldPerfOutput(job *job.ProfilingJob, pid string) (error, string)
	handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string) error
	publishResult(c compressor.Type, fileName string, outputType api.OutputType) error
}

type perfManager struct {
}

func NewPerfProfiler() *PerfProfiler {
	return &PerfProfiler{
		delay:       perfDelayBetweenJobs,
		PerfManager: &perfManager{},
	}
}

var perfCommander = exec.NewCommander()

func (p *PerfProfiler) SetUp(job *job.ProfilingJob) error {
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

func (p *PerfProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	files := make([]string, 0, len(p.targetPIDs))

	pool := pond.New(len(p.targetPIDs), 0, pond.MinWorkers(len(p.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range p.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, f, _ := p.invoke(job, pid)
			if err == nil {
				files = append(files, f)
			}
			return err
		})
		// wait a bit between jobs for not overloading the system (bcc-profiler is a heavy tool)
		time.Sleep(p.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()
	if err != nil {
		return err, time.Since(start)
	}

	fileName := common.GetResultFile(common.TmpDir(), job.Tool, api.Raw)
	// merge all files
	file.MergeFiles(fileName, files)

	err = p.handleProfilingResult(job, flamegraph.Get(job), fileName)
	if err != nil {
		return err, time.Since(start)
	}

	return p.publishResult(job.Compressor, common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType), job.OutputType), time.Since(start)
}

func (m *perfManager) invoke(job *job.ProfilingJob, pid string) (error, string, time.Duration) {
	start := time.Now()
	err := m.runPerfRecord(job, pid)
	if err != nil {
		return errors.Wrap(err, "perf record failed"), "", time.Since(start)
	}

	err = m.runPerfScript(job, pid)
	if err != nil {
		return errors.Wrap(err, "perf script failed"), "", time.Since(start)
	}

	err, fileName := m.foldPerfOutput(job, pid)
	if err != nil {
		return errors.Wrap(err, "folding perf output failed"), "", time.Since(start)
	}

	return nil, fileName, time.Since(start)
}

func (m *perfManager) runPerfRecord(job *job.ProfilingJob, pid string) error {
	duration := strconv.Itoa(int(job.Duration.Seconds()))
	var stderr bytes.Buffer
	cmd := perfCommander.Command(perfLocation, "record", "-p", pid, "-o", fmt.Sprintf(perfRecordOutputFileName, pid), "-g", "--", "sleep", duration)
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}
	return err
}

func (m *perfManager) runPerfScript(job *job.ProfilingJob, pid string) error {
	f, err := os.Create(fmt.Sprintf(perfScriptOutputFileName, pid))
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("error closing resource: %s", err)
			return
		}
	}(f)

	var stderr bytes.Buffer
	cmd := perfCommander.Command(perfLocation, "script", "-i", fmt.Sprintf(perfRecordOutputFileName, pid))
	cmd.Stdout = f
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}
	return err
}

func (m *perfManager) foldPerfOutput(job *job.ProfilingJob, pid string) (error, string) {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := perfCommander.Command(flameGraphStackCollapseLocation, fmt.Sprintf(perfScriptOutputFileName, pid))
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), ""
	}

	// out file name is composed by the job info and the pid
	fileName := common.GetResultFile(common.TmpDir(), job.Tool, api.Raw) + "." + pid
	// add process pid legend to each line of the output and write it to the file
	file.Write(fileName, addProcessPIDLegend(out.String(), pid))

	return err, fileName
}

func (m *perfManager) handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string) error {
	if job.OutputType == api.FlameGraph {
		if file.GetSize(fileName) < common.MinimumRawSize {
			return fmt.Errorf("unable to generate flamegraph: no stacks found (maybe due low cpu load)")
		}
		// convert raw format to flamegraph
		err := flameGrapher.StackSamplesToFlameGraph(fileName, common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType))
		if err != nil {
			return errors.Wrap(err, "could not convert raw format to flamegraph")
		}
	}
	return nil
}

func (m *perfManager) publishResult(c compressor.Type, fileName string, outputType api.OutputType) error {
	return publish.Do(c, fileName, outputType)
}

func (p *PerfProfiler) CleanUp(job *job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), "perf")
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix+string(job.OutputType))

	return nil
}
