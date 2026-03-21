package profiler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	dotnetAppDir           = "/app"
)

// getTargetTmpDir generates the path to the target process's temporary directory using its process ID (pid).
var getTargetTmpDir = func(pid string) string {
	return fmt.Sprintf("/proc/%s/root/tmp", pid)
}

// getInnerPID reads the /proc/<pid>/status file to find the innermost process ID (innerPID) of the target process.
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

// findDiagnosticSocket finds the diagnostic socket for a given process ID (pid) and its innermost process ID (innerPID).
// It reads the temporary directory of the target process and looks for a socket file that matches the expected naming pattern.
// If found, it returns the name of the socket file; otherwise, it returns an empty string.
var findDiagnosticSocket = func(pid string, innerPID string) string {
	targetTmpDir := getTargetTmpDir(pid)
	files, err := os.ReadDir(targetTmpDir)
	if err != nil {
		return ""
	}
	prefix := fmt.Sprintf("dotnet-diagnostic-%s-", innerPID)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), prefix) && strings.HasSuffix(f.Name(), "-socket") {
			return f.Name()
		}
	}
	return ""
}

// setTmpDir sets the TMPDIR environment variable to the target process's temporary directory.'
var setTmpDir = func(cmd *exec.Cmd, pid string) error {
	innerPID, err := getInnerPID(pid)
	if err != nil {
		innerPID = pid
	}

	if innerPID != pid {
		socketName := findDiagnosticSocket(pid, innerPID)
		if socketName != "" {
			// Extract the key from the socket name
			key := strings.TrimPrefix(socketName, fmt.Sprintf("dotnet-diagnostic-%s-", innerPID))
			hostSocketName := fmt.Sprintf("dotnet-diagnostic-%s-%s", pid, key)
			hostSocketPath := filepath.Join("/tmp", hostSocketName)
			targetSocketPath := filepath.Join(getTargetTmpDir(pid), socketName)

			_ = os.Remove(hostSocketPath)
			err := os.Symlink(targetSocketPath, hostSocketPath)
			if err != nil {
				log.ErrorLogLn(fmt.Sprintf("could not create socket symlink: %s", err))
			}
		}
	}

	cmd.Env = append(os.Environ(), "TMPDIR=/tmp")
	return nil
}

// formatDotnetDuration converts a time.Duration to the hh:mm:ss format required by dotnet-trace.
func formatDotnetDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// dotnetTraceCommand creates a command to invoke dotnet-trace for CPU profiling.
var dotnetTraceCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	duration := formatDotnetDuration(job.Interval)
	traceLocation := filepath.Join(dotnetAppDir, "dotnet-trace")

	args := []string{"collect", "-p", pid, "--duration", duration, "-o", fileName}
	if job.OutputType == api.SpeedScope {
		args = append(args, "--format", "Speedscope")
	}
	cmd := commander.Command(traceLocation, args...)
	_ = setTmpDir(cmd, pid)
	return cmd
}

// dotnetGcdumpCommand creates a command to invoke dotnet-gcdump for memory/GC heap analysis.
var dotnetGcdumpCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	gcdumpLocation := filepath.Join(dotnetAppDir, "dotnet-gcdump")

	args := []string{"collect", "-v", "-p", pid, "-o", fileName}
	cmd := commander.Command(gcdumpLocation, args...)
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
		log.ErrorLogLn(stderr.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	// When dotnet-trace converts to Speedscope format it produces a NEW file with the
	// ".speedscope.json" suffix (e.g. "agent-speedscope-3504-1.json" ->
	// "agent-speedscope-3504-1.speedscope.json").
	if job.Tool == api.DotnetTrace && job.OutputType == api.SpeedScope {
		resultFileName = strings.TrimSuffix(resultFileName, filepath.Ext(resultFileName)) + ".speedscope.json"
	}

	return p.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (p *DotnetProfiler) CleanUp(*job.ProfilingJob) error {
	for _, pid := range p.targetPIDs {
		matches, _ := filepath.Glob(fmt.Sprintf("/tmp/dotnet-diagnostic-%s-*-socket", pid))
		for _, m := range matches {
			_ = os.Remove(m)
		}
	}
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
