package profiler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/alitto/pond"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
)

const (
	perfLocation                    = "/app/perf"
	perfRecordOutputFileName        = "/tmp/perf-%s-%d.data"
	flameGraphStackCollapseLocation = "/app/FlameGraph/stackcollapse-perf.pl"
	perfScriptOutputFileName        = "/tmp/perf-%s-%d.out"
	perfDelayBetweenJobs            = 2 * time.Second
)

type PerfProfiler struct {
	targetPIDs []string
	delay      time.Duration
	PerfManager
}

type PerfManager interface {
	invoke(job *job.ProfilingJob, pid string) (error, time.Duration)
	runPerfRecord(job *job.ProfilingJob, pid string) error
	runPerfScript(job *job.ProfilingJob, pid string) error
	foldPerfOutput(job *job.ProfilingJob, pid string) (error, string)
	handleFlamegraph(*job.ProfilingJob, flamegraph.FrameGrapher, string, string) error
}

type perfManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

func NewPerfProfiler(commander executil.Commander, publisher publish.Publisher) *PerfProfiler {
	return &PerfProfiler{
		delay: perfDelayBetweenJobs,
		PerfManager: &perfManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

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

	pool := pond.New(len(p.targetPIDs), 0, pond.MinWorkers(len(p.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range p.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, _ := p.invoke(job, pid)
			return err
		})
		// wait a bit between jobs for not overloading the system
		time.Sleep(p.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()

	return err, time.Since(start)
}

func (m *perfManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()
	err := m.runPerfRecord(job, pid)
	if err != nil {
		return errors.Wrap(err, "perf record failed"), time.Since(start)
	}

	err = m.runPerfScript(job, pid)
	if err != nil {
		return errors.Wrap(err, "perf script failed"), time.Since(start)
	}

	err, fileName := m.foldPerfOutput(job, pid)
	if err != nil {
		return errors.Wrap(err, "folding perf output failed"), time.Since(start)
	}

	// out file names is composed by the job info and the pid
	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)

	err = m.handleFlamegraph(job, flamegraph.Get(job), fileName, resultFileName)
	if err != nil {
		log.ErrorLogLn(fmt.Sprintf("could not generate flamegraph (PID: %s): %s", pid, err.Error()))
		return nil, time.Since(start)
	}

	return m.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (m *perfManager) runPerfRecord(job *job.ProfilingJob, pid string) error {
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	var stderr bytes.Buffer
	cmd := m.commander.Command(perfLocation, "record", "-p", pid, "-o", fmt.Sprintf(perfRecordOutputFileName, pid, job.Iteration), "-g", "--", "sleep", interval)
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}
	return err
}

func (m *perfManager) runPerfScript(job *job.ProfilingJob, pid string) error {
	f, err := os.Create(fmt.Sprintf(perfScriptOutputFileName, pid, job.Iteration))
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
	cmd := m.commander.Command(perfLocation, "script", "-i", fmt.Sprintf(perfRecordOutputFileName, pid, job.Iteration))
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

	cmd := m.commander.Command(flameGraphStackCollapseLocation, fmt.Sprintf(perfScriptOutputFileName, pid, job.Iteration))
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), ""
	}

	// out file name is composed by the job info and the pid
	fileName := common.GetResultFile(common.TmpDir(), job.Tool, api.Raw, pid, job.Iteration)
	// add process pid legend to each line of the output and write it to the file
	file.Write(fileName, addProcessPIDLegend(out.String(), pid))

	return err, fileName
}

func (m *perfManager) handleFlamegraph(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, rawFileName string,
	flameFileName string) error {
	if job.OutputType == api.FlameGraph {
		if file.Size(rawFileName) < common.MinimumRawSize {
			return fmt.Errorf("unable to generate flamegraph: no stacks found (maybe due low cpu load)")
		}
		// convert raw format to flamegraph
		err := flameGrapher.StackSamplesToFlameGraph(rawFileName, flameFileName)
		if err != nil {
			return errors.Wrap(err, "could not convert raw format to flamegraph")
		}
	}
	return nil
}

func (p *PerfProfiler) CleanUp(job *job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), "perf")
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix+string(job.OutputType))

	return nil
}
