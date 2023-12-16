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
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"os/exec"
	"strconv"
	"time"
)

const (
	rbSpyLocation = "/app/rbspy"
)

var rubyCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	output := job.OutputType

	interval := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{"record"}
	args = append(args, "--pid", pid, "--file", fileName, "--duration", interval, "--format", string(output))
	return util.Command(rbSpyLocation, args...)

}

type RubyProfiler struct {
	targetPID string
	cmd       *exec.Cmd
	RubyManager
}

type RubyManager interface {
	publishResult(compressor compressor.Type, fileName string, outputType api.OutputType) error
	cleanUp(cmd *exec.Cmd)
}

type rubyManager struct {
}

func NewRubyProfiler() *RubyProfiler {
	return &RubyProfiler{
		RubyManager: &rubyManager{},
	}
}

func (r *RubyProfiler) SetUp(job *job.ProfilingJob) error {
	if stringUtils.IsNotBlank(job.PID) {
		r.targetPID = job.PID
	} else {
		pid, err := util.ContainerPID(job)
		if err != nil {
			return err
		}
		r.targetPID = pid
	}
	log.DebugLogLn(fmt.Sprintf("The PID to be profiled: %s", r.targetPID))

	return nil
}

func (r *RubyProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	fileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType)

	var out bytes.Buffer
	var stderr bytes.Buffer

	r.cmd = rubyCommand(job, r.targetPID, fileName)
	r.cmd.Stdout = &out
	r.cmd.Stderr = &stderr
	err := r.cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	return r.publishResult(job.Compressor, common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType), job.OutputType), time.Since(start)
}

func (r *RubyProfiler) CleanUp(job *job.ProfilingJob) error {
	r.cleanUp(r.cmd)

	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}

func (p *rubyManager) publishResult(c compressor.Type, fileName string, outputType api.OutputType) error {
	return util.Publish(c, fileName, outputType)
}

func (p *rubyManager) cleanUp(cmd *exec.Cmd) {
	if cmd != nil && cmd.ProcessState == nil {
		err := cmd.Process.Kill()
		if err != nil {
			log.WarningLogLn(fmt.Sprintf("unable kill process: %s", err))
		}
	}
}
