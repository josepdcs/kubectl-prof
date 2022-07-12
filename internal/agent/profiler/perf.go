package profiler

import (
	"bytes"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	utils2 "github.com/josepdcs/kubectl-prof/internal/agent/utils"
	"os"
	"strconv"
)

const (
	perfLocation                    = "/app/perf"
	perfRecordOutputFileName        = "/tmp/perf.data"
	flameGraphPlLocation            = "/app/FlameGraph/flamegraph.pl"
	flameGraphStackCollapseLocation = "/app/FlameGraph/stackcollapse-perf.pl"
	perfScriptOutputFileName        = "/tmp/perf.out"
	perfFoldedOutputFileName        = "/tmp/perf.folded"
)

var perfResultFile = func(job *config.ProfilingJob) string {
	if stringUtils.IsBlank(job.FileName) {
		return "/tmp/" + job.FileName
	}
	return "/tmp/flamegraph.svg"
}

type PerfProfiler struct {
	PerfUtil
}

type PerfUtil interface {
	runPerfRecord(job *config.ProfilingJob) error
	runPerfScript(job *config.ProfilingJob) error
	foldPerfOutput(job *config.ProfilingJob) error
	generateFlameGraph(job *config.ProfilingJob) error
	publishResult(compressor api.Compressor, fileName string, outputType api.EventType) error
}

type perfUtil struct {
}

func NewPerfProfiler() *PerfProfiler {
	return &PerfProfiler{&perfUtil{}}
}

func (p *PerfProfiler) SetUp(job *config.ProfilingJob) error {
	return nil
}

func (p *PerfProfiler) Invoke(job *config.ProfilingJob) error {
	err := p.runPerfRecord(job)
	if err != nil {
		return fmt.Errorf("perf record failed: %s", err)
	}

	err = p.runPerfScript(job)
	if err != nil {
		return fmt.Errorf("perf script failed: %s", err)
	}

	err = p.foldPerfOutput(job)
	if err != nil {
		return fmt.Errorf("folding perf output failed: %s", err)
	}

	err = p.generateFlameGraph(job)
	if err != nil {
		return fmt.Errorf("flamegraph generation failed: %s", err)
	}

	return p.publishResult(job.Compressor, perfResultFile(job), job.OutputType)
}

func (p *PerfProfiler) CleanUp(job *config.ProfilingJob) error {
	fileName := perfResultFile(job)
	err := os.Remove(fileName + api.GetExtensionFileByCompressor[job.Compressor])
	if err != nil {
		utils2.PublishLogEvent(api.WarnLevel, fmt.Sprintf("file could no be removed: %s", err))
	}
	return os.Remove(fileName)
}

func (p *perfUtil) runPerfRecord(job *config.ProfilingJob) error {
	pid, err := utils2.ContainerPID(job, true)
	if err != nil {
		return err
	}
	utils2.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))

	var stderr bytes.Buffer
	cmd := utils2.Command(perfLocation, "record", "-p", pid, "-o", perfRecordOutputFileName, "-g", "--", "sleep", duration)
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		utils2.PublishLogEvent(api.ErrorLevel, stderr.String())
	}
	return err
}

func (p *perfUtil) runPerfScript(job *config.ProfilingJob) error {
	f, err := os.Create(perfScriptOutputFileName)
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

	var stderr bytes.Buffer
	cmd := utils2.Command(perfLocation, "script", "-i", perfRecordOutputFileName)
	cmd.Stdout = f
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		utils2.PublishLogEvent(api.ErrorLevel, stderr.String())
	}
	return err
}

func (p *perfUtil) foldPerfOutput(job *config.ProfilingJob) error {
	f, err := os.Create(perfFoldedOutputFileName)
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

	var stderr bytes.Buffer
	cmd := utils2.Command(flameGraphStackCollapseLocation, perfScriptOutputFileName)
	cmd.Stdout = f
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		utils2.PublishLogEvent(api.ErrorLevel, stderr.String())
	}
	return err
}

func (p *perfUtil) generateFlameGraph(job *config.ProfilingJob) error {
	inputFile, err := os.Open(perfFoldedOutputFileName)
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

	outputFile, err := os.Create(perfResultFile(job))
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

	var stderr bytes.Buffer
	cmd := utils2.Command(flameGraphPlLocation, "--colors", getColors(job.Language))
	cmd.Stdin = inputFile
	cmd.Stdout = outputFile
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		utils2.PublishLogEvent(api.ErrorLevel, stderr.String())
	}
	return err
}

func (b *perfUtil) publishResult(c api.Compressor, fileName string, outputType api.EventType) error {
	return utils2.Publish(c, fileName, outputType)
}

func getColors(l api.ProgrammingLanguage) string {
	switch l {
	case api.Node:
		return "js"
	case api.Java:
		return "java"
	default:
		return "green"
	}
}
