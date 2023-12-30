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
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
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

var bpfCommander = executil.NewCommander()

var bccProfilerCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{"-df", "-U", "-p", pid, interval}
	return bpfCommander.Command(profilerLocation, args...)
}

type BpfProfiler struct {
	targetPIDs []string
	delay      time.Duration
	BpfManager
}

type BpfManager interface {
	invoke(job *job.ProfilingJob, pid string) (error, string, time.Duration)
	handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string) error
	publishResult(compressor compressor.Type, fileName string, outputType api.OutputType) error
}

type bpfManager struct {
}

func NewBpfProfiler() *BpfProfiler {
	return &BpfProfiler{
		delay:      bpfDelayBetweenJobs,
		BpfManager: &bpfManager{},
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

	files := make([]string, 0, len(b.targetPIDs))

	pool := pond.New(len(b.targetPIDs), 0, pond.MinWorkers(len(b.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range b.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, f, _ := b.invoke(job, pid)
			if err == nil {
				files = append(files, f)
			}
			return err
		})
		// wait a bit between jobs for not overloading the system (bcc-profiler is a heavy tool)
		time.Sleep(b.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()
	if err != nil {
		return err, time.Since(start)
	}

	fileName := common.GetResultFile(common.TmpDir(), job.Tool, api.Raw)
	// merge all files
	file.MergeFiles(fileName, files)

	err = b.handleProfilingResult(job, flamegraph.Get(job), fileName)
	if err != nil {
		return err, time.Since(start)
	}

	return b.publishResult(job.Compressor, common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType), job.OutputType), time.Since(start)
}

func (b *bpfManager) invoke(job *job.ProfilingJob, pid string) (error, string, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := bccProfilerCommand(job, pid)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), "", time.Since(start)
	}

	// out file name is composed by the job info and the pid
	fileName := common.GetResultFile(common.TmpDir(), job.Tool, api.Raw) + "." + pid
	// add process pid legend to each line of the output and write it to the file
	file.Write(fileName, addProcessPIDLegend(out.String(), pid))

	return nil, fileName, time.Since(start)
}

func (b *bpfManager) handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string) error {
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

func (b *bpfManager) publishResult(c compressor.Type, fileName string, outputType api.OutputType) error {
	return util.Publish(c, fileName, outputType)
}

func (b *BpfProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
