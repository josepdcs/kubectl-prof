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
	dotnetTraceLocation       = "/app/dotnet-trace"
	dotnetGcdumpLocation      = "/app/dotnet-gcdump"
	dotnetDelayBetweenJobs    = 2 * time.Second
)

var dotnetTraceCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	duration := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{"collect", "-p", pid, "--duration", fmt.Sprintf("00:00:%s", duration), "-o", fileName}
	if job.OutputType == api.SpeedScope {
		args = append(args, "--format", "Speedscope")
	}
	return commander.Command(dotnetTraceLocation, args...)
}

var dotnetGcdumpCommand = func(commander executil.Commander, pid string, fileName string) *exec.Cmd {
	args := []string{"collect", "-p", pid, "-o", fileName}
	return commander.Command(dotnetGcdumpLocation, args...)
}

// DotnetProfiler implements the Profiler interface for .NET Core/5+ applications using
// dotnet-trace (CPU profiling) and dotnet-gcdump (memory/GC heap analysis).
type DotnetProfiler struct {
	targetPIDs []string
	delay      time.Duration
	DotnetManager
}

// DotnetManager defines the interface for .NET profiling operations.
type DotnetManager interface {
	invoke(job *job.ProfilingJob, pid string) (error, time.Duration)
}

type dotnetManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

// NewDotnetProfiler creates a new DotnetProfiler instance.
func NewDotnetProfiler(commander executil.Commander, publisher publish.Publisher) *DotnetProfiler {
	return &DotnetProfiler{
		delay: dotnetDelayBetweenJobs,
		DotnetManager: &dotnetManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

func (p *DotnetProfiler) SetUp(job *job.ProfilingJob) error {
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

func (p *DotnetProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
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

func (p *dotnetManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)

	var cmd *exec.Cmd
	switch job.Tool {
	case api.DotnetGcdump:
		cmd = dotnetGcdumpCommand(p.commander, pid, resultFileName)
	default:
		// api.DotnetTrace
		cmd = dotnetTraceCommand(p.commander, job, pid, resultFileName)
	}

	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	return p.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (p *DotnetProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
