package profiler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
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
	args := []string{"-p", pid, "-o", fileName, "--palette", "rust", "--title", fmt.Sprintf("Flamegraph for PID %s", pid)}
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

	// Set up the process group so we can send signals to the entire process tree
	// This is crucial because flamegraph spawns perf as a subprocess
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
	}

	// Get current environment variables
	currentEnv := os.Environ()
	// Force LC_ALL to C for consistent locale behavior
	currentEnv = append(currentEnv, "LC_ALL=C")
	cmd.Env = currentEnv
	cmd.Dir = "/tmp"
	// Ensure no stale perf.data exists
	file.RemoveAll(common.TmpDir(), "perf.data")

	// Start the profiler process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start profiler: %w", err), time.Since(start)
	}

	// Create a timer to send SIGINT (Ctrl+C equivalent) after the specified interval
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

	// Give flamegraph time to flush the output file after SIGTERM
	// Flamegraph needs to process perf data and write the SVG after receiving the signal
	log.DebugLogLn(fmt.Sprintf("Waiting for flamegraph to complete file writing: %s", fileName))

	// Poll for the file existence with timeout
	maxWait := 5 * time.Second
	pollInterval := 500 * time.Millisecond
	waited := time.Duration(0)

	for waited < maxWait {
		if file.Exists(fileName) {
			log.DebugLogLn(fmt.Sprintf("Output file found after %v: %s", waited, fileName))
			break
		}
		time.Sleep(pollInterval)
		waited += pollInterval
	}

	// Verify the output file exists before trying to publish
	if !file.Exists(fileName) {
		log.ErrorLogLn(fmt.Sprintf("Output file not found after waiting %v: %s", waited, fileName))

		// Log stdout and stderr for debugging
		if out.Len() > 0 {
			log.DebugLogLn(fmt.Sprintf("flamegraph stdout: %s", out.String()))
		}
		if stderr.Len() > 0 {
			log.DebugLogLn(fmt.Sprintf("flamegraph stderr: %s", stderr.String()))
		}

		return fmt.Errorf("output file not found: %s", fileName), time.Since(start)
	}

	return p.publisher.Do(job.Compressor, fileName, job.OutputType), time.Since(start)
}

// sendTerminationSignal sends SIGINT to the flamegraph process
// SIGINT (Ctrl+C) allows flamegraph and perf to terminate gracefully
func (p *rustManager) sendTerminationSignal(cmd *exec.Cmd, interval time.Duration) {
	if cmd.Process != nil {
		log.DebugLogLn(fmt.Sprintf("Sending SIGINT to flamegraph process (PID: %d) after %v", cmd.Process.Pid, interval))

		// Try sending to process group first
		pgid := cmd.Process.Pid
		if err := syscall.Kill(-pgid, syscall.SIGINT); err != nil {
			log.DebugLogLn(fmt.Sprintf("Could not send to process group, sending to process directly: %v", err))
			// Fallback: send to just the process
			if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
				log.ErrorLogLn(fmt.Sprintf("Failed to send SIGINT to process %d: %v", cmd.Process.Pid, err))
			}
		}
	}
}

// isExpectedTermination checks if the error is due to SIGINT we sent (expected behavior)
func (p *rustManager) isExpectedTermination(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}

	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		return false
	}

	// If the process was terminated by SIGINT (Ctrl+C), it's expected
	if status.Signaled() && status.Signal() == syscall.SIGINT {
		log.DebugLogLn("Flamegraph process terminated successfully with SIGINT")
		return true
	}

	return false
}

func (r *RustProfiler) CleanUp(job *job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
