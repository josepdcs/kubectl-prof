package profiler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/alitto/pond"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
)

const (
	cargoFlameLocation         = "/app/flamegraph"
	cargoFlameDelayBetweenJobs = 2 * time.Second
)

var rustCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	args := []string{"-p", pid, "-o", fileName, "--root"}
	return commander.Command(cargoFlameLocation, args...)
}

type RustProfiler struct {
	targetPIDs []string
	delay      time.Duration
	RustManager
}

type RustManager interface {
	invoke(job *job.ProfilingJob, pid string) (error, time.Duration)
}

type rustManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

func NewRustProfiler(commander executil.Commander, publisher publish.Publisher) *RustProfiler {
	return &RustProfiler{
		delay: cargoFlameDelayBetweenJobs,
		RustManager: &rustManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

func (r *RustProfiler) SetUp(job *job.ProfilingJob) error {
	if stringUtils.IsNotBlank(job.PID) {
		r.targetPIDs = []string{job.PID}
		return nil
	}
	pids, err := util.GetCandidatePIDs(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PIDs to be profiled: %s", pids))
	r.targetPIDs = pids

	return nil
}

func (r *RustProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	// create a pool of workers
	pool := pond.New(len(r.targetPIDs), 0, pond.MinWorkers(len(r.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range r.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, _ := r.invoke(job, pid)
			return err
		})
		// wait a bit between jobs for not overloading the system
		time.Sleep(r.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()

	return err, time.Since(start)
}

func (p *rustManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	fileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)
	cmd := rustCommand(p.commander, job, pid, fileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Start the profiler process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start profiler: %w", err), time.Since(start)
	}

	log.DebugLogLn(fmt.Sprintf("Flamegraph process started (PID: %d), interval: %s", cmd.Process.Pid, job.Interval.String()))

	// Create a timer to send SIGTERM after the specified interval
	timer := time.AfterFunc(job.Interval, func() {
		p.sendTerminationSignal(cmd, job.Interval)
	})
	defer timer.Stop()

	// Wait for the process to finish
	err := cmd.Wait()
	if err != nil && !p.isExpectedTermination(err) {
		// Log both stdout and stderr for debugging
		if out.Len() > 0 {
			log.ErrorLogLn(fmt.Sprintf("stdout: %s", out.String()))
		}
		if stderr.Len() > 0 {
			log.ErrorLogLn(fmt.Sprintf("stderr: %s", stderr.String()))
		}
		return fmt.Errorf("could not launch profiler: %w - stderr: %s", err, stderr.String()), time.Since(start)
	}

	return p.publisher.Do(job.Compressor, fileName, job.OutputType), time.Since(start)
}

// sendTerminationSignal sends SIGTERM to the flamegraph process
func (p *rustManager) sendTerminationSignal(cmd *exec.Cmd, interval time.Duration) {
	if cmd.Process != nil {
		log.DebugLogLn(fmt.Sprintf("Sending SIGTERM to flamegraph process (PID: %d) after %v", cmd.Process.Pid, interval))
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.ErrorLogLn(fmt.Sprintf("Failed to send SIGTERM to flamegraph process: %v", err))
		}
	}
}

// isExpectedTermination checks if the error is due to SIGTERM we sent (expected behavior)
func (p *rustManager) isExpectedTermination(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}

	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		return false
	}

	// If the process was terminated by SIGTERM, it's expected
	if status.Signaled() && status.Signal() == syscall.SIGTERM {
		log.DebugLogLn("Flamegraph process terminated successfully with SIGTERM")
		return true
	}

	return false
}

func (r *RustProfiler) CleanUp(job *job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
