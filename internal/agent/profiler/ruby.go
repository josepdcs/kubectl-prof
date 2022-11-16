package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"strconv"
	"time"
)

const (
	rbspyLocation = "/app/rbspy"
)

type RubyProfiler struct{}

func NewRubyProfiler() *RubyProfiler {
	return &RubyProfiler{}
}

func (r *RubyProfiler) SetUp(job *job.ProfilingJob) error {
	return nil
}

func (r *RubyProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()
	pid, err := util.ContainerPID(job)
	if err != nil {
		return err, time.Since(start)
	}
	log.DebugLogLn(fmt.Sprintf("The PID to be profiled: %s", pid))

	filName := common.GetResultFile(common.TmpDir(), job)
	duration := strconv.Itoa(int(job.Duration.Seconds()))
	cmd := util.Command(rbspyLocation, "record", "--pid", pid, "--file", filName, "--duration", duration, "--format", "flamegraph")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.ErrorLogLn(stderr.String())
		return fmt.Errorf("could not launch profiler: %w", err), time.Since(start)
	}

	return util.Publish(job.Compressor, filName, job.OutputType), time.Since(start)
}

func (r *RubyProfiler) CleanUp(job *job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
