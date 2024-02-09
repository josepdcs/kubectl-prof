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
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"os/exec"
	"strconv"
	"time"
)

const (
	profilerLocation    = "/app/bcc-profiler/profile"
	bpfDelayBetweenJobs = 5 * time.Second
)

var bccProfilerCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string) *exec.Cmd {
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{"-df", "-U", "-F", "99", "-p", pid, interval}
	return commander.Command(profilerLocation, args...)
}

type BpfProfiler struct {
	targetPIDs []string
	delay      time.Duration
	BpfManager
}

type BpfManager interface {
	invoke(*job.ProfilingJob, string) (error, time.Duration)
	handleFlamegraph(*job.ProfilingJob, flamegraph.FrameGrapher, string, string) error
}

type bpfManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

func NewBpfProfiler(commander executil.Commander, publisher publish.Publisher) *BpfProfiler {
	return &BpfProfiler{
		delay: bpfDelayBetweenJobs,
		BpfManager: &bpfManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

func (b *BpfProfiler) SetUp(job *job.ProfilingJob) error {
	if stringUtils.IsNotBlank(job.PID) {
		b.targetPIDs = []string{job.PID}
		return nil
	}
	pids, err := util.GetCandidatePIDs(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PIDs to be profiled: %s", pids))
	b.targetPIDs = pids

	return nil
}

func (b *BpfProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	pool := pond.New(len(b.targetPIDs), 0, pond.MinWorkers(len(b.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range b.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, _ := b.invoke(job, pid)
			return err
		})
		// wait a bit between jobs for not overloading the system
		time.Sleep(b.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()

	return err, time.Since(start)
}

func (b *bpfManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := bccProfilerCommand(b.commander, job, pid)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	// out file names is composed by the job info and the pid
	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)
	fileName := common.GetResultFile(common.TmpDir(), job.Tool, api.Raw, pid, job.Iteration)
	// add process pid legend to each line of the output and write it to the file
	file.Write(fileName, addProcessPIDLegend(out.String(), pid))

	err = b.handleFlamegraph(job, flamegraph.Get(job), fileName, resultFileName)
	if err != nil {
		log.ErrorLogLn(fmt.Sprintf("could not generate flamegraph (PID: %s): %s", pid, err.Error()))
		return nil, time.Since(start)
	}

	return b.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (b *bpfManager) handleFlamegraph(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, rawFileName string,
	flameFileName string) error {
	if job.OutputType == api.FlameGraph {
		if file.GetSize(rawFileName) < common.MinimumRawSize {
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

func (b *BpfProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
