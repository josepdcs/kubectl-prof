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
	"strconv"
	"time"
)

const (
	perfLocation                    = "/app/perf"
	perfRecordOutputFileName        = "/tmp/perf.data"
	flameGraphPlLocation            = "/app/FlameGraph/flamegraph.pl"
	flameGraphStackCollapseLocation = "/app/FlameGraph/stackcollapse-perf.pl"
	perfScriptOutputFileName        = "/tmp/perf.out"
	perfFoldedOutputFileName        = "/tmp/perf.folded"
)

type PerfProfiler struct {
	PerfUtil
}

type PerfUtil interface {
	runPerfRecord(job *job.ProfilingJob) error
	runPerfScript(job *job.ProfilingJob) error
	foldPerfOutput(job *job.ProfilingJob) error
	generateFlameGraph(job *job.ProfilingJob) error
	publishResult(c compressor.Type, fileName string, outputType api.EventType) error
}

type perfUtil struct {
}

func NewPerfProfiler() *PerfProfiler {
	return &PerfProfiler{&perfUtil{}}
}

func (p *PerfProfiler) SetUp(job *job.ProfilingJob) error {
	return nil
}

func (p *PerfProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()
	err := p.runPerfRecord(job)
	if err != nil {
		return fmt.Errorf("perf record failed: %s", err), time.Since(start)
	}

	err = p.runPerfScript(job)
	if err != nil {
		return fmt.Errorf("perf script failed: %s", err), time.Since(start)
	}

	err = p.foldPerfOutput(job)
	if err != nil {
		return fmt.Errorf("folding perf output failed: %s", err), time.Since(start)
	}

	err = p.generateFlameGraph(job)
	if err != nil {
		return fmt.Errorf("flamegraph generation failed: %s", err), time.Since(start)
	}

	return p.publishResult(job.Compressor, common.GetResultFile(common.TmpDir(), job), job.OutputType), time.Since(start)
}

func (p *PerfProfiler) CleanUp(job *job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix+string(job.OutputType))

	return nil
}

func (b *perfUtil) runPerfRecord(job *job.ProfilingJob) error {
	pid, err := util.ContainerPID(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))

	var stderr bytes.Buffer
	cmd := util.Command(perfLocation, "record", "-p", pid, "-o", perfRecordOutputFileName, "-g", "--", "sleep", duration)
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}
	return err
}

func (b *perfUtil) runPerfScript(job *job.ProfilingJob) error {
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
	cmd := util.Command(perfLocation, "script", "-i", perfRecordOutputFileName)
	cmd.Stdout = f
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}
	return err
}

func (b *perfUtil) foldPerfOutput(job *job.ProfilingJob) error {
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
	cmd := util.Command(flameGraphStackCollapseLocation, perfScriptOutputFileName)
	cmd.Stdout = f
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}
	return err
}

func (b *perfUtil) generateFlameGraph(job *job.ProfilingJob) error {
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

	outputFile, err := os.Create(common.GetResultFile(common.TmpDir(), job))
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
	cmd := util.Command(flameGraphPlLocation, "--colors", getColors(job.Language))
	cmd.Stdin = inputFile
	cmd.Stdout = outputFile
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
	}
	return err
}

func (b *perfUtil) publishResult(c compressor.Type, fileName string, outputType api.EventType) error {
	return util.Publish(c, fileName, outputType)
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
