package profiler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/alitto/pond"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"os/exec"
	"strconv"
	"time"
)

const (
	rbSpyLocation         = "/app/rbspy"
	rbSpyDelayBetweenJobs = 2 * time.Second
)

var rubyCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	output := job.OutputType

	interval := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{"record"}
	args = append(args, "--pid", pid, "--file", fileName, "--duration", interval, "--format", string(output))
	return commander.Command(rbSpyLocation, args...)

}

type RubyProfiler struct {
	targetPIDs []string
	delay      time.Duration
	RubyManager
}

type RubyManager interface {
	invoke(job *job.ProfilingJob, pid string) (error, time.Duration)
}

type rubyManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

func NewRubyProfiler(commander executil.Commander, publisher publish.Publisher) *RubyProfiler {
	return &RubyProfiler{
		delay: rbSpyDelayBetweenJobs,
		RubyManager: &rubyManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

func (r *RubyProfiler) SetUp(job *job.ProfilingJob) error {
	if stringUtils.IsNotBlank(job.PID) {
		r.targetPIDs = []string{job.PID}
		return nil
	}
	pids, err := util.GetCandidatePIDs(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PIDs to be profiled: %s", pids))
	r.targetPIDs = pids

	return nil
}

func (r *RubyProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	// create a pool of workers
	pool := pond.New(len(r.targetPIDs), 0, pond.MinWorkers(len(r.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range r.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, _ := r.invoke(job, pid)
			return err
		})
		// wait a bit between jobs for not overloading the system
		time.Sleep(r.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()

	return err, time.Since(start)
}

func (p *rubyManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	fileName := common.GetResultFileWithPID(common.TmpDir(), job.Tool, job.OutputType, pid)
	cmd := rubyCommand(p.commander, job, pid, fileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	return p.publisher.Do(job.Compressor, fileName, job.OutputType), time.Since(start)
}

func (r *RubyProfiler) CleanUp(job *job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
