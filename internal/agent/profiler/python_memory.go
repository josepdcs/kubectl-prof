package profiler

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
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
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
)

const (
	memrayLocation         = "/app/memray"
	memrayDelayBetweenJobs = 2 * time.Second
)

var memrayCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, rawFileName string) *exec.Cmd {
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{"attach", "--aggregate", "-o", rawFileName, "--duration", interval, pid}
	return commander.Command(memrayLocation, args...)
}

var memrayReportCommand = func(commander executil.Commander, job *job.ProfilingJob, rawFileName string, outputFileName string) *exec.Cmd {
	switch job.OutputType {
	case api.FlameGraph:
		args := []string{"flamegraph", rawFileName, "-o", outputFileName}
		return commander.Command(memrayLocation, args...)
	default:
		// api.Summary or api.Tree — output goes to stdout
		args := []string{string(job.OutputType), rawFileName}
		return commander.Command(memrayLocation, args...)
	}
}

// MemrayProfiler profiles Python processes using Memray memory profiler.
type MemrayProfiler struct {
	targetPIDs []string
	delay      time.Duration
	MemrayManager
}

// MemrayManager abstracts the inner profiling operations so they can be mocked in tests.
type MemrayManager interface {
	invoke(*job.ProfilingJob, string) (error, time.Duration)
	handleReport(*job.ProfilingJob, string, string) error
}

type memrayManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

func NewMemrayProfiler(commander executil.Commander, publisher publish.Publisher) *MemrayProfiler {
	return &MemrayProfiler{
		delay: memrayDelayBetweenJobs,
		MemrayManager: &memrayManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

func (p *MemrayProfiler) SetUp(job *job.ProfilingJob) error {
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

func (p *MemrayProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
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

func (p *memrayManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	// intermediate raw binary file
	rawFileName := common.GetResultFile(common.TmpDir(), job.Tool, api.Raw, pid, job.Iteration)

	cmd := memrayCommand(p.commander, job, pid, rawFileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)
	err = p.handleReport(job, rawFileName, resultFileName)
	if err != nil {
		log.ErrorLogLn(fmt.Sprintf("could not generate report (PID: %s): %s", pid, err.Error()))
		return nil, time.Since(start)
	}

	return p.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (p *memrayManager) handleReport(job *job.ProfilingJob, rawFileName string, resultFileName string) error {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := memrayReportCommand(p.commander, job, rawFileName, resultFileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "could not generate %s report: %s", string(job.OutputType), stderr.String())
	}

	if job.OutputType != api.FlameGraph {
		// for summary/tree the report writes to stdout; persist it to the result file
		file.Write(resultFileName, out.String())
	}

	return nil
}

func (p *MemrayProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
