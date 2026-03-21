package profiler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/alitto/pond"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
)

const (
	dotnetDelayBetweenJobs = 2 * time.Second
)

var dotnetAppDir = "/app"

var dotnetTmpDirSymlink = "/tmp/k-prof-dotnet"

var getTargetTmpDir = func(pid string) string {
	return fmt.Sprintf("/proc/%s/root/tmp", pid)
}

var getInnerPID = func(pid string) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%s/status", pid))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "NSpid:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				// The last field is the PID in the innermost namespace
				return fields[len(fields)-1], nil
			}
		}
	}
	return pid, nil
}

var setTmpDir = func(cmd *exec.Cmd, pid string) error {
	targetTmpDir := getTargetTmpDir(pid)
	_ = os.Remove(dotnetTmpDirSymlink)
	err := os.Symlink(targetTmpDir, dotnetTmpDirSymlink)
	if err != nil {
		log.ErrorLogLn(fmt.Sprintf("could not create symlink: %s", err))
		// fallback to the long path if symlink creation fails
		cmd.Env = append(os.Environ(), fmt.Sprintf("TMPDIR=%s", targetTmpDir))
		return nil
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("TMPDIR=%s", dotnetTmpDirSymlink))
	return nil
}

var dotnetTraceCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	duration := strconv.Itoa(int(job.Interval.Seconds()))
	traceLocation := filepath.Join(dotnetAppDir, "dotnet-trace")
	innerPID, err := getInnerPID(pid)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("could not get inner PID for %s: %s", pid, err))
		innerPID = pid
	}

	args := []string{"-t", pid, "-p", "--", traceLocation, "collect", "-p", innerPID, "--duration", fmt.Sprintf("00:00:%s", duration), "-o", fileName}
	if job.OutputType == api.SpeedScope {
		args = append(args, "--format", "Speedscope")
	}
	cmd := commander.Command("nsenter", args...)
	_ = setTmpDir(cmd, pid)
	return cmd
}

var dotnetGcdumpCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	gcdumpLocation := filepath.Join(dotnetAppDir, "dotnet-gcdump")
	innerPID, err := getInnerPID(pid)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("could not get inner PID for %s: %s", pid, err))
		innerPID = pid
	}

	args := []string{"-t", pid, "-p", "--", gcdumpLocation, "collect", "-p", innerPID, "-o", fileName}
	cmd := commander.Command("nsenter", args...)
	_ = setTmpDir(cmd, pid)
	return cmd
}

// DotnetProfiler implements the Profiler interface for .NET Core/5+ applications using
// dotnet-trace (CPU profiling) and dotnet-gcdump (memory/GC heap analysis).
type DotnetProfiler struct {
	targetPIDs []string
	delay      time.Duration
	DotnetManager
}

// DotnetManager defines the interface for .NET profiling operations.
type DotnetManager interface {
	invoke(job *job.ProfilingJob, pid string) (error, time.Duration)
}

type dotnetManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

// NewDotnetProfiler creates a new DotnetProfiler instance.
func NewDotnetProfiler(commander executil.Commander, publisher publish.Publisher) *DotnetProfiler {
	return &DotnetProfiler{
		delay: dotnetDelayBetweenJobs,
		DotnetManager: &dotnetManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

func (p *DotnetProfiler) SetUp(job *job.ProfilingJob) error {
	if stringUtils.IsNotBlank(job.PID) {
		p.targetPIDs = []string{job.PID}
	} else {
		pids, err := util.GetCandidatePIDs(job)
		if err != nil {
			return err
		}
		log.DebugLogLn(fmt.Sprintf("The PIDs to be profiled: %s", pids))
		p.targetPIDs = pids
	}

	for _, pid := range p.targetPIDs {
		_ = pid
	}

	return nil
}

func (p *DotnetProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	pool := pond.New(len(p.targetPIDs), 0, pond.MinWorkers(len(p.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range p.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, _ := p.invoke(job, pid)
			return err
		})
		// wait a bit between jobs for not overloading the system
		time.Sleep(p.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()

	return err, time.Since(start)
}

func (p *dotnetManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)

	var cmd *exec.Cmd
	switch job.Tool {
	case api.DotnetGcdump:
		cmd = dotnetGcdumpCommand(p.commander, job, pid, resultFileName)
	default:
		// api.DotnetTrace
		cmd = dotnetTraceCommand(p.commander, job, pid, resultFileName)
	}

	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	return p.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (p *DotnetProfiler) CleanUp(*job.ProfilingJob) error {
	_ = os.Remove(dotnetTmpDirSymlink)
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
