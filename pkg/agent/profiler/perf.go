package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
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
	flameGraphPerfOutputFile        = "/tmp/perf.svg"
)

type PerfProfiler struct{}

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

	return utils.Publish(job.Compressor, flameGraphPerfOutputFile, job.OutputType)
}

func (p *PerfProfiler) runPerfRecord(job *config.ProfilingJob) error {
	pid, err := utils.ContainerPID(job, true)
	if err != nil {
		return err
	}
	api.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))

	//cmd := utils.Command(perfLocation, "record", "-F", "99", "-p", pid, "-o", perfRecordOutputFileName, "-g", "--", "sleep", duration)
	var stderr bytes.Buffer
	cmd := utils.Command(perfLocation, "record", "-p", pid, "-o", perfRecordOutputFileName, "-g", "--", "sleep", duration)
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		api.PublishLogEvent(api.ErrorLevel, stderr.String())
	}
	return err
}

func (p *PerfProfiler) runPerfScript(job *config.ProfilingJob) error {
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
	cmd := utils.Command(perfLocation, "script", "-i", perfRecordOutputFileName)
	cmd.Stdout = f
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		api.PublishLogEvent(api.ErrorLevel, stderr.String())
	}
	return err
}

func (p *PerfProfiler) foldPerfOutput(job *config.ProfilingJob) error {
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
	cmd := utils.Command(flameGraphStackCollapseLocation, perfScriptOutputFileName)
	cmd.Stdout = f
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		api.PublishLogEvent(api.ErrorLevel, stderr.String())
	}
	return err
}

func (p *PerfProfiler) generateFlameGraph(job *config.ProfilingJob) error {
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

	outputFile, err := os.Create(flameGraphPerfOutputFile)
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
	cmd := utils.Command(flameGraphPlLocation, "--colors", getColors(job.Language))
	cmd.Stdin = inputFile
	cmd.Stdout = outputFile
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		api.PublishLogEvent(api.ErrorLevel, stderr.String())
	}
	return err
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
