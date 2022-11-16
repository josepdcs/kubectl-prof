package profiler

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
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	profilerLocation         = "/app/bcc-profiler/profile"
	rawProfilerOutputFile    = "/tmp/raw_profile.txt"
	flameGraphScriptLocation = "/app/FlameGraph/flamegraph.pl"
)

type BpfProfiler struct {
	targetPID string
	BpfManager
}

type BpfManager interface {
	runProfiler(*job.ProfilingJob, string) error
	generateFlameGraph(fileName string) error
	publishResult(compressor compressor.Type, fileName string, outputType api.EventType) error
	cleanUp()
}

type bpfManager struct {
	profCmd  *exec.Cmd
	flameCmd *exec.Cmd
}

func NewBpfProfiler() *BpfProfiler {
	return &BpfProfiler{BpfManager: &bpfManager{}}
}

func (b *BpfProfiler) SetUp(job *job.ProfilingJob) error {
	pid, err := util.ContainerPID(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PID to be profiled: %s", pid))
	b.targetPID = pid

	return nil
}

func (b *BpfProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()
	err := b.runProfiler(job, b.targetPID)
	if err != nil {
		return fmt.Errorf("profiling failed: %s", err), time.Since(start)
	}

	fileName := common.GetResultFile(common.TmpDir(), job)
	err = b.generateFlameGraph(fileName)
	if err != nil {
		return fmt.Errorf("flamegraph generation failed: %s", err), time.Since(start)
	}

	return b.publishResult(job.Compressor, fileName, job.OutputType), time.Since(start)
}

func (b *BpfProfiler) CleanUp(*job.ProfilingJob) error {
	b.cleanUp()

	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}

func (b *bpfManager) runProfiler(job *job.ProfilingJob, targetPID string) error {
	f, err := os.Create(rawProfilerOutputFile)
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

	interval := strconv.Itoa(int(job.Interval.Seconds()))
	var stderr bytes.Buffer
	args := []string{"-df", "-U", "-p", targetPID, interval}
	b.profCmd = util.Command(profilerLocation, args...)
	b.profCmd.Stdout = f
	b.profCmd.Stderr = &stderr

	err = b.profCmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}
	return err
}

func (b *bpfManager) generateFlameGraph(fileName string) error {
	inputFile, err := os.Open(rawProfilerOutputFile)
	if err != nil {
		return err
	}

	defer func(inputFile *os.File) {
		err := inputFile.Close()
		if err != nil {
			log.ErrorLogLn(fmt.Sprintf("error closing input file: %s", err))
			return
		}
	}(inputFile)

	outputFile, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			log.ErrorLogLn(fmt.Sprintf("error closing output file: %s", err))
			return
		}
	}(outputFile)

	var stderr bytes.Buffer
	b.flameCmd = util.Command(flameGraphScriptLocation)
	b.flameCmd.Stdin = inputFile
	b.flameCmd.Stdout = outputFile
	b.flameCmd.Stderr = &stderr

	err = b.flameCmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}

	return err
}

func (b *bpfManager) publishResult(c compressor.Type, fileName string, outputType api.EventType) error {
	return util.Publish(c, fileName, outputType)
}

func (b *bpfManager) cleanUp() {
	if b.profCmd != nil && b.profCmd.ProcessState == nil {
		err := b.profCmd.Process.Kill()
		if err != nil {
			log.WarningLogLn(fmt.Sprintf("unable kill process: %s", err))
		} else {
			log.DebugLogLn("try to kill prof process")
		}
	}
	if b.flameCmd != nil && b.flameCmd.ProcessState == nil {
		err := b.flameCmd.Process.Kill()
		if err != nil {
			log.WarningLogLn(fmt.Sprintf("unable kill process: %s", err))
		} else {
			log.DebugLogLn("try to kill flamegraph generator process")
		}
	}
}
