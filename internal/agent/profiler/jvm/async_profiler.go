package jvm

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const (
	asyncProfilerDir = "/tmp/async-profiler"
	profilerSh       = asyncProfilerDir + "/profiler.sh"
)

var asyncProfilerCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	event := string(job.Event)
	output := string(job.OutputType)
	args := []string{"-o", output, "-d", interval, "-f", fileName, "-e", event, "--fdtransfer", pid}
	return util.Command(profilerSh, args...)
}

var asyncProfilerStopCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
	return util.Command(profilerSh, "stop", pid)
}

type AsyncProfiler struct {
	targetPID string
	AsyncProfilerManager
}

type AsyncProfilerManager interface {
	removeTmpDir() error
	linkTmpDirToTargetTmpDir(string) error
	copyProfilerToTmpDir() error
	publishResult(compressor compressor.Type, fileName string, outputType api.EventType) error
}

type asyncProfilerManager struct {
}

func NewAsyncProfiler() *AsyncProfiler {
	return &AsyncProfiler{AsyncProfilerManager: &asyncProfilerManager{}}
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

	pid, err := util.ContainerPID(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PID to be profiled: %s", pid))
	j.targetPID = pid

	return j.copyProfilerToTmpDir()
}

func (j *AsyncProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()
	var out bytes.Buffer
	var stderr bytes.Buffer

	fileName := common.GetResultFile(common.TmpDir(), job)
	cmd := asyncProfilerCommand(job, j.targetPID, fileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		log.ErrorLogLn(stderr.String())
		return fmt.Errorf("could not launch profiler: %w; detail: %s", err, stderr.String()), time.Since(start)
	}
	log.DebugLogLn(out.String())

	return j.publishResult(job.Compressor, fileName, job.OutputType), time.Since(start)
}

func (j *AsyncProfiler) CleanUp(job *job.ProfilingJob) error {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := asyncProfilerStopCommand(job, j.targetPID)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.WarningLogLn(stderr.String())
	}
	_, _ = fmt.Fprint(io.Discard, out.String())

	err = os.RemoveAll(asyncProfilerDir)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("async-profiler folder could not be removed: %s", err))
	}
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix+string(job.OutputType))

	return nil
}

func (j *asyncProfilerManager) removeTmpDir() error {
	return os.RemoveAll(common.TmpDir())
}

func (j *asyncProfilerManager) linkTmpDirToTargetTmpDir(targetTmpDir string) error {
	return os.Symlink(targetTmpDir, common.TmpDir())
}

func (j *asyncProfilerManager) copyProfilerToTmpDir() error {
	cmd := util.Command("cp", "-r", "/app/async-profiler", common.TmpDir())
	return cmd.Run()
}

func (j *asyncProfilerManager) publishResult(c compressor.Type, fileName string, outputType api.EventType) error {
	return util.Publish(c, fileName, outputType)
}
