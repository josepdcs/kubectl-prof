package profiler

import (
	"bytes"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	kernelSourcesDir         = "/usr/src/kernel-source/"
	profilerLocation         = "/app/bcc-profiler/profile"
	rawProfilerOutputFile    = "/tmp/raw_profile.txt"
	flameGraphScriptLocation = "/app/FlameGraph/flamegraph.pl"
)

var bpfResultFile = func(job *config.ProfilingJob) string {
	if stringUtils.IsBlank(job.FileName) {
		return "/tmp/" + job.FileName
	}
	return "/tmp/flamegraph.svg"
}

type BpfProfiler struct {
	BpfUtil
}

type BpfUtil interface {
	runProfiler(job *config.ProfilingJob) error
	generateFlameGraph(fileName string) error
	moveSources(target string) error
	publishResult(compressor api.Compressor, fileName string, outputType api.EventType) error
}

type bpfUtil struct {
}

func NewBpfProfiler() *BpfProfiler {
	return &BpfProfiler{&bpfUtil{}}
}

func (b *BpfProfiler) SetUp(job *config.ProfilingJob) error {
	exitCode, kernelVersion, err := utils.ExecuteCommand(utils.Command("uname", "-r"))
	if err != nil {
		return fmt.Errorf("failed to get kernel version, exit code: %d, error: %s", exitCode, err)
	}

	expectedSourcesLocation, err := os.Readlink(fmt.Sprintf("/lib/modules/%s/build",
		strings.TrimSuffix(kernelVersion, "\n")))
	if err != nil {
		return fmt.Errorf("failed to read source link, error: %s", err)
	}

	return b.moveSources(expectedSourcesLocation)
}

func (b *BpfProfiler) Invoke(job *config.ProfilingJob) error {
	err := b.runProfiler(job)
	if err != nil {
		return fmt.Errorf("profiling failed: %s", err)
	}

	fileName := bpfResultFile(job)
	err = b.generateFlameGraph(fileName)
	if err != nil {
		return fmt.Errorf("flamegraph generation failed: %s", err)
	}

	return b.publishResult(job.Compressor, fileName, job.OutputType)
}

func (b *BpfProfiler) CleanUp(job *config.ProfilingJob) error {
	fileName := bpfResultFile(job)
	err := os.Remove(fileName + api.GetExtensionFileByCompressor[job.Compressor])
	if err != nil {
		utils.PublishLogEvent(api.WarnLevel, fmt.Sprintf("file could no be removed: %s", err))
	}
	return os.Remove(fileName)
}

func (b *bpfUtil) moveSources(target string) error {
	parent, _ := filepath.Split(target)
	err := os.MkdirAll(parent, os.ModePerm)
	if err != nil {
		return err
	}

	_, _, err = utils.ExecuteCommand(utils.Command("mv", kernelSourcesDir, target))
	if err != nil {
		return fmt.Errorf("failed moving source files, error: %s, tried to move to: %s", err, target)
	}

	return nil
}

func (b *bpfUtil) runProfiler(job *config.ProfilingJob) error {
	pid, err := utils.ContainerPID(job, false)
	if err != nil {
		return err
	}
	utils.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

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

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	var stderr bytes.Buffer
	cmd := utils.Command(profilerLocation, "-df", "-U", "-p", pid, duration)
	cmd.Stdout = f
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		utils.PublishLogEvent(api.ErrorLevel, stderr.String())
	}
	return err
}

func (b *bpfUtil) generateFlameGraph(fileName string) error {
	inputFile, err := os.Open(rawProfilerOutputFile)
	if err != nil {
		return err
	}

	defer func(inputFile *os.File) {
		err := inputFile.Close()
		if err != nil {
			fmt.Printf("error closing input file: %s", err)
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
			fmt.Printf("error closing output file: %s", err)
			return
		}
	}(outputFile)

	flameGraphCmd := exec.Command(flameGraphScriptLocation)
	flameGraphCmd.Stdin = inputFile
	flameGraphCmd.Stdout = outputFile

	return flameGraphCmd.Run()
}

func (b *bpfUtil) publishResult(c api.Compressor, fileName string, outputType api.EventType) error {
	return utils.Publish(c, fileName, outputType)
}
