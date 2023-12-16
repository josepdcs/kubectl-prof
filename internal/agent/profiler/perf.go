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
	"strconv"
	"time"
)

const (
	perfLocation                    = "/app/perf"
	perfRecordOutputFileName        = "/tmp/perf.data"
	flameGraphStackCollapseLocation = "/app/FlameGraph/stackcollapse-perf.pl"
	perfScriptOutputFileName        = "/tmp/perf.out"
)

type PerfProfiler struct {
	PerfUtil
}

type PerfUtil interface {
	runPerfRecord(job *job.ProfilingJob) error
	runPerfScript(job *job.ProfilingJob) error
	foldPerfOutput(job *job.ProfilingJob) error
	generateFlameGraph(job *job.ProfilingJob) error
	publishResult(c compressor.Type, fileName string, outputType api.OutputType) error
}

type perfUtil struct {
}

func NewPerfProfiler() *PerfProfiler {
	return &PerfProfiler{&perfUtil{}}
}

func (p *PerfProfiler) SetUp(*job.ProfilingJob) error {
	return nil
}

func (p *PerfProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()
	err := p.runPerfRecord(job)
	if err != nil {
		return errors.Wrap(err, "perf record failed"), time.Since(start)
	}

	err = p.runPerfScript(job)
	if err != nil {
		return errors.Wrap(err, "perf script failed"), time.Since(start)
	}

	err = p.foldPerfOutput(job)
	if err != nil {
		return errors.Wrap(err, "folding perf output failed"), time.Since(start)
	}

	if job.OutputType == api.FlameGraph {
		err = p.generateFlameGraph(job)
		if err != nil {
			return errors.Wrap(err, "flamegraph generation failed"), time.Since(start)
		}
	}

	return p.publishResult(job.Compressor, common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType), job.OutputType), time.Since(start)
}

func (p *PerfProfiler) CleanUp(job *job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix+string(job.OutputType))

	return nil
}

func (b *perfUtil) runPerfRecord(job *job.ProfilingJob) error {
	var pid string
	var err error
	if stringUtils.IsNotBlank(job.PID) {
		pid = job.PID
	} else {
		pid, err = util.ContainerPID(job)
		if err != nil {
			return err
		}
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

func (b *perfUtil) runPerfScript(*job.ProfilingJob) error {
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
	f, err := os.Create(common.GetResultFile(common.TmpDir(), job.Tool, api.Raw))
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
	flameGrapher := flamegraph.Get(job)
	// convert raw format to flamegraph
	err := flameGrapher.StackSamplesToFlameGraph(common.GetResultFile(common.TmpDir(), job.Tool, api.Raw),
		common.GetResultFile(common.TmpDir(), job.Tool, api.FlameGraph))
	if err != nil {
		return errors.Wrap(err, "could not convert raw format to flamegraph")
	}
	return nil
}

func (b *perfUtil) publishResult(c compressor.Type, fileName string, outputType api.OutputType) error {
	return util.Publish(c, fileName, outputType)
}
