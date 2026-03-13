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
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
)

const (
	phpSpyLocation         = "/app/phpspy"
	phpSpyDelayBetweenJobs = 2 * time.Second
)

var phpspyCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	duration := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{"-p", pid, "-T", duration, "-f", "flamegraph", "-o", fileName}
	return commander.Command(phpSpyLocation, args...)
}

// PhpspyProfiler implements the Profiler interface for PHP applications using phpspy.
type PhpspyProfiler struct {
	targetPIDs []string
	delay      time.Duration
	PhpspyManager
}

// PhpspyManager defines the interface for phpspy profiling operations.
type PhpspyManager interface {
	invoke(job *job.ProfilingJob, pid string) (error, time.Duration)
	handleFlamegraph(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, rawFileName string, flameFileName string) error
}

type phpspyManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

// NewPhpspyProfiler creates a new PhpspyProfiler instance.
func NewPhpspyProfiler(commander executil.Commander, publisher publish.Publisher) *PhpspyProfiler {
	return &PhpspyProfiler{
		delay: phpSpyDelayBetweenJobs,
		PhpspyManager: &phpspyManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

func (p *PhpspyProfiler) SetUp(job *job.ProfilingJob) error {
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

func (p *PhpspyProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
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

func (p *phpspyManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	fileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)
	if job.OutputType == api.FlameGraph {
		fileName = common.GetResultFile(common.TmpDir(), job.Tool, api.Raw, pid, job.Iteration)
	}
	cmd := phpspyCommand(p.commander, job, pid, fileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	// result file name is composed by the job info and the pid
	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)
	err = p.handleFlamegraph(job, flamegraph.Get(job), fileName, resultFileName)
	if err != nil {
		log.ErrorLogLn(fmt.Sprintf("could not generate flamegraph (PID: %s): %s", pid, err.Error()))
		return nil, time.Since(start)
	}

	return p.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (p *phpspyManager) handleFlamegraph(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher,
	rawFileName string, flameFileName string) error {
	if job.OutputType == api.FlameGraph {
		if file.IsEmpty(rawFileName) {
			return errors.New("unable to generate flamegraph: no stacks found (maybe due low cpu load)")
		}
		// convert raw format to flamegraph
		err := flameGrapher.StackSamplesToFlameGraph(rawFileName, flameFileName)
		if err != nil {
			return errors.Wrap(err, "could not convert raw format to flamegraph")
		}
	}
	return nil
}

func (p *PhpspyProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
