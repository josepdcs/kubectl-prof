package profiler

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
	"os"
	"os/exec"
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

	return utils.PublishFlameGraph(flameGraphPerfOutputFile)
}

func (p *PerfProfiler) runPerfRecord(job *config.ProfilingJob) error {
	pid, err := utils.ContainerPID(job, true)
	if err != nil {
		return err
	}
	api.PublishLogEvent(api.InfoLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	cmd := exec.Command(perfLocation, "record", "-p", pid, "-o", perfRecordOutputFileName, "-g", "--", "sleep", duration)

	return cmd.Run()
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

	cmd := exec.Command(perfLocation, "script", "-i", perfRecordOutputFileName)
	cmd.Stdout = f

	return cmd.Run()
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

	cmd := exec.Command(flameGraphStackCollapseLocation, perfScriptOutputFileName)
	cmd.Stdout = f

	return cmd.Run()
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

	cmd := exec.Command(flameGraphPlLocation)
	cmd.Stdin = inputFile
	cmd.Stdout = outputFile

	return cmd.Run()
}
