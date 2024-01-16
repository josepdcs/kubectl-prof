package jvm

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
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const (
	asyncProfilerDir              = "/tmp/async-profiler"
	profilerSh                    = asyncProfilerDir + "/profiler.sh"
	asyncProfilerDelayBetweenJobs = 2 * time.Second
)

var asyncProfilerCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	event := string(job.Event)
	output := string(job.OutputType)
	if job.OutputType == api.Raw {
		// overrides to collapsed type since it is the type defined be async-profiler, which it is what we want
		output = string(api.Collapsed)
	}
	args := []string{"-o", output, "-d", interval, "-f", fileName, "-e", event, "--fdtransfer", pid}
	return commander.Command(profilerSh, args...)
}

var asyncProfilerStopCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string) *exec.Cmd {
	return commander.Command(profilerSh, "stop", pid)
}

type AsyncProfiler struct {
	targetPIDs []string
	delay      time.Duration
	AsyncProfilerManager
}

type AsyncProfilerManager interface {
	removeTmpDir() error
	linkTmpDirToTargetTmpDir(string) error
	copyProfilerToTmpDir() error
	invoke(*job.ProfilingJob, string) (error, time.Duration)
	cleanUp(*job.ProfilingJob, string)
}

type asyncProfilerManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

func NewAsyncProfiler(commander executil.Commander, publisher publish.Publisher) *AsyncProfiler {
	return &AsyncProfiler{
		delay: asyncProfilerDelayBetweenJobs,
		AsyncProfilerManager: &asyncProfilerManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

func (j *AsyncProfiler) SetUp(job *job.ProfilingJob) error {
	targetFs, err := util.ContainerFileSystem(job.ContainerRuntime, job.ContainerID)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The target filesystem is: %s", targetFs))

	err = j.removeTmpDir()
	if err != nil {
		return err
	}

	targetTmpDir := filepath.Join(targetFs, "tmp")
	// remove previous files from a before profiling
	file.RemoveAll(targetTmpDir, config.ProfilingPrefix+string(job.OutputType))

	err = j.linkTmpDirToTargetTmpDir(targetTmpDir)
	if err != nil {
		return err
	}

	if stringUtils.IsNotBlank(job.PID) {
		j.targetPIDs = []string{job.PID}
	} else {
		pids, err := util.GetCandidatePIDs(job)
		if err != nil {
			return err
		}
		log.DebugLogLn(fmt.Sprintf("The PIDs to be profiled: %s", pids))
		j.targetPIDs = pids
	}

	return j.copyProfilerToTmpDir()
}

func (j *asyncProfilerManager) removeTmpDir() error {
	return os.RemoveAll(common.TmpDir())
}

func (j *asyncProfilerManager) linkTmpDirToTargetTmpDir(targetTmpDir string) error {
	return os.Symlink(targetTmpDir, common.TmpDir())
}

func (j *asyncProfilerManager) copyProfilerToTmpDir() error {
	cmd := j.commander.Command("cp", "-r", "/app/async-profiler", common.TmpDir())
	return cmd.Run()
}

func (j *AsyncProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	pool := pond.New(len(j.targetPIDs), 0, pond.MinWorkers(len(j.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range j.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, _ := j.invoke(job, pid)
			return err
		})
		// wait a bit between jobs for not overloading the system
		time.Sleep(j.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()

	return err, time.Since(start)
}

func (j *asyncProfilerManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()
	var out bytes.Buffer
	var stderr bytes.Buffer

	resultFileName := common.GetResultFileWithPID(common.TmpDir(), job.Tool, job.OutputType, pid)
	cmd := asyncProfilerCommand(j.commander, job, pid, resultFileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}
	log.DebugLogLn(out.String())

	return j.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (j *AsyncProfiler) CleanUp(job *job.ProfilingJob) error {
	for _, pid := range j.targetPIDs {
		j.cleanUp(job, pid)
	}

	err := os.RemoveAll(asyncProfilerDir)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("async-profiler folder could not be removed: %s", err))
	}
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix+string(job.OutputType))

	return nil
}

func (j *asyncProfilerManager) cleanUp(job *job.ProfilingJob, pid string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := asyncProfilerStopCommand(j.commander, job, pid)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.WarningLogLn(stderr.String())
	}
	_, _ = fmt.Fprint(io.Discard, out.String())
}
