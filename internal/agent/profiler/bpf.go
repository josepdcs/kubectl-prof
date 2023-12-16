package profiler

import (
	"bytes"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	profilerLocation = "/app/bcc-profiler/profile"
)

var bccProfilerCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{"-df", "-U", "-p", pid, interval}
	return util.Command(profilerLocation, args...)
}

type BpfProfiler struct {
	targetPID string
	cmd       *exec.Cmd
	BpfManager
}

type BpfManager interface {
	handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string, out bytes.Buffer) error
	publishResult(compressor compressor.Type, fileName string, outputType api.OutputType) error
	cleanUp(cmd *exec.Cmd)
}

type bpfManager struct {
}

func NewBpfProfiler() *BpfProfiler {
	return &BpfProfiler{BpfManager: &bpfManager{}}
}

func (b *BpfProfiler) SetUp(job *job.ProfilingJob) error {
	if stringUtils.IsNotBlank(job.PID) {
		b.targetPID = job.PID
	} else {
		pid, err := util.ContainerPID(job)
		if err != nil {
			return err
		}
		b.targetPID = pid
	}
	log.DebugLogLn(fmt.Sprintf("The PID to be profiled: %s", b.targetPID))

	return nil
}

func (b *BpfProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()
	var out bytes.Buffer
	var stderr bytes.Buffer

	b.cmd = bccProfilerCommand(job, b.targetPID)
	b.cmd.Stdout = &out
	b.cmd.Stderr = &stderr
	err := b.cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	err = b.handleProfilingResult(job, flamegraph.Get(job), common.GetResultFile(common.TmpDir(), job.Tool, api.Raw), out)
	if err != nil {
		return err, time.Since(start)
	}

	return b.publishResult(job.Compressor, common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType), job.OutputType), time.Since(start)
}

func (b *BpfProfiler) CleanUp(*job.ProfilingJob) error {
	b.cleanUp(b.cmd)

	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}

func (b *bpfManager) handleProfilingResult(job *job.ProfilingJob, flameGrapher flamegraph.FrameGrapher, fileName string, out bytes.Buffer) error {
	err := os.WriteFile(fileName, out.Bytes(), 0644)
	if err != nil {
		return errors.Wrapf(err, "could not save file: %s", fileName)
	}
	if job.OutputType == api.FlameGraph {
		if stringUtils.IsBlank(out.String()) {
			return errors.New("unable to generate flamegraph: no stacks found (maybe due low cpu load)")
		}
		// convert raw format to flamegraph
		err = flameGrapher.StackSamplesToFlameGraph(fileName, common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType))
		if err != nil {
			return errors.Wrap(err, "could not convert raw format to flamegraph")
		}
	}
	return nil
}

func (b *bpfManager) publishResult(c compressor.Type, fileName string, outputType api.OutputType) error {
	return util.Publish(c, fileName, outputType)
}

func (b *bpfManager) cleanUp(cmd *exec.Cmd) {
	if cmd != nil && cmd.ProcessState == nil {
		err := cmd.Process.Kill()
		if err != nil {
			log.WarningLogLn(fmt.Sprintf("unable kill process: %s", err))
		}
	}
}
