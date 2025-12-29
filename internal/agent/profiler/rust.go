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
	args := []string{"-p", pid, "-o", fileName, "--verbose"}
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

	// Log environment for debugging
	log.DebugLogLn(fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
	log.DebugLogLn(fmt.Sprintf("PERF=%s", os.Getenv("PERF")))

	// Verify perf is accessible
	perfCheckCmd := p.commander.Command("which", "perf")
	if output, err := perfCheckCmd.CombinedOutput(); err != nil {
		log.ErrorLogLn(fmt.Sprintf("perf not found in PATH: %v, output: %s", err, string(output)))

		// Try absolute path
		perfCheckCmd2 := p.commander.Command("ls", "-la", "/app/perf")
		if output2, err2 := perfCheckCmd2.CombinedOutput(); err2 != nil {
			log.ErrorLogLn(fmt.Sprintf("/app/perf not accessible: %v, output: %s", err2, string(output2)))
		} else {
			log.DebugLogLn(fmt.Sprintf("/app/perf exists: %s", string(output2)))
		}
	} else {
		log.DebugLogLn(fmt.Sprintf("perf found at: %s", string(output)))
	}

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

	// Ensure PATH and PERF are set correctly for the subprocess
	currentEnv := os.Environ()
	pathSet := false
	perfSet := false
	for i, env := range currentEnv {
		if len(env) > 5 && env[:5] == "PATH=" {
			// Ensure /app and /usr/bin are in PATH
			currentEnv[i] = "PATH=/app:/usr/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/sbin:/bin"
			pathSet = true
		}
		if len(env) > 5 && env[:5] == "PERF=" {
			currentEnv[i] = "PERF=/usr/bin/perf"
			perfSet = true
		}
	}
	if !pathSet {
		currentEnv = append(currentEnv, "PATH=/app:/usr/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/sbin:/bin")
	}
	if !perfSet {
		currentEnv = append(currentEnv, "PERF=/usr/bin/perf")
	}
	// Forzar locale C para evitar problemas de UTF-8 en la salida de perf
	currentEnv = append(currentEnv, "LC_ALL=C")
	cmd.Env = currentEnv
	cmd.Dir = "/tmp"
	// Ensure no stale perf.data exists
	_ = os.Remove("/tmp/perf.data")

	log.DebugLogLn(fmt.Sprintf("Command: %s %v", cmd.Path, cmd.Args))

	// Start the profiler process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start profiler: %w", err), time.Since(start)
	}

	log.DebugLogLn(fmt.Sprintf("Flamegraph process started (PID: %d), interval: %s", cmd.Process.Pid, job.Interval.String()))

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
	maxWait := 30 * time.Second
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

		// List files in tmp directory for debugging
		if entries, err := os.ReadDir(common.TmpDir()); err == nil {
			log.DebugLogLn(fmt.Sprintf("Files in %s:", common.TmpDir()))
			for _, entry := range entries {
				log.DebugLogLn(fmt.Sprintf("  - %s", entry.Name()))
			}
		}
		return fmt.Errorf("output file not found: %s", fileName), time.Since(start)
	}

	log.DebugLogLn(fmt.Sprintf("Output file successfully created: %s", fileName))
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
