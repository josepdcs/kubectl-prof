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
	memrayLocation         = "/app/memray"
	memrayDelayBetweenJobs = 2 * time.Second
)

// memrayCommand runs memray attach --aggregate inside the target's network namespace via nsenter.
// --aggregate makes the tracker write directly to a file (no TCP back-channel), which avoids
// gevent's socket monkey-patching breaking the connection.
var memrayCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, rawFileName string) *exec.Cmd {
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	netNs := fmt.Sprintf("/proc/%s/ns/net", pid)
	args := []string{"--net=" + netNs, "--", memrayLocation, "attach", "--aggregate", "--verbose", "-o", rawFileName, "--duration", interval, pid}
	return commander.Command("nsenter", args...)
}

// stageMemrayLib copies memray's _inject.abi3.so from the profiling container into the target
// container's filesystem at its exact same path so that gdb's dlopen() call (which runs in the
// target's mount namespace) can find our version of the library.
// Only _inject.abi3.so is staged; _memray.cpython-312.so is intentionally left as the target's
// native version so the tracker writes a file in the target's format (read back via nsenter).
// Returns a cleanup function that restores the original file (or removes the staged one).
var stageMemrayLib = func(pid string) (func(), error) {
	out, err := exec.Command("python3", "-c",
		"import memray, pathlib\n"+
			"print(pathlib.Path(memray.__file__).parent / '_inject.abi3.so')").Output()
	if err != nil {
		return nil, fmt.Errorf("could not locate memray inject library: %w", err)
	}

	libSrc := strings.TrimSpace(string(out))
	if libSrc == "" {
		return nil, fmt.Errorf("could not locate memray inject library: empty output")
	}
	libDst := fmt.Sprintf("/proc/%s/root%s", pid, libSrc)

	if err := os.MkdirAll(filepath.Dir(libDst), 0755); err != nil {
		return nil, fmt.Errorf("could not create directory for %s: %w", libDst, err)
	}

	data, err := os.ReadFile(libSrc)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", libSrc, err)
	}

	// Save whatever the target had (if anything) for restoration on cleanup.
	oldData, _ := os.ReadFile(libDst)
	hadOld := oldData != nil

	if err := os.WriteFile(libDst, data, 0755); err != nil {
		return nil, fmt.Errorf("could not stage %s to %s: %w", libSrc, libDst, err)
	}

	return func() {
		if hadOld {
			_ = os.WriteFile(libDst, oldData, 0755)
		} else {
			_ = os.Remove(libDst)
		}
	}, nil
}

var memrayReportCommand = func(commander executil.Commander, job *job.ProfilingJob, rawFileName string, outputFileName string) *exec.Cmd {
	switch job.OutputType {
	case api.FlameGraph:
		args := []string{"flamegraph", rawFileName, "-o", outputFileName}
		return commander.Command(memrayLocation, args...)
	case api.Summary:
		args := []string{"summary", rawFileName}
		return commander.Command(memrayLocation, args...)
	case api.Tree:
		args := []string{"tree", rawFileName}
		return commander.Command(memrayLocation, args...)
	default:
		return nil
	}
}

// MemrayProfiler profiles Python processes using Memray memory profiler.
type MemrayProfiler struct {
	targetPIDs []string
	delay      time.Duration
	MemrayManager
}

// MemrayManager abstracts the inner profiling operations so they can be mocked in tests.
type MemrayManager interface {
	invoke(*job.ProfilingJob, string) (error, time.Duration)
	handleReport(*job.ProfilingJob, string, string) error
}

type memrayManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

func NewMemrayProfiler(commander executil.Commander, publisher publish.Publisher) *MemrayProfiler {
	return &MemrayProfiler{
		delay: memrayDelayBetweenJobs,
		MemrayManager: &memrayManager{
			commander: commander,
			publisher: publisher,
		},
	}
}

func (p *MemrayProfiler) SetUp(job *job.ProfilingJob) error {
	if stringUtils.IsNotBlank(job.PID) {
		p.targetPIDs = []string{job.PID}
		return nil
	}
	pids, err := util.GetCandidatePIDs(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PIDs to be profiled: %s", pids))
	p.targetPIDs = pids

	return nil
}

func (p *MemrayProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
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

func (p *memrayManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	// Stage _inject.abi3.so into the target container's filesystem so that gdb's dlopen()
	// call (which runs in the target's mount namespace) can find the library.
	cleanup, err := stageMemrayLib(pid)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("could not stage memray inject library for PID %s: %s (profiling may fail)", pid, err))
	} else {
		defer cleanup()
	}

	// intermediate raw binary file
	rawFileName := common.GetResultFile(common.TmpDir(), job.Tool, api.Raw, pid, job.Iteration)

	// The tracker (--aggregate) writes the raw file to its own /tmp (target's mount namespace).
	// Clean both locations before running to avoid "file exists" errors from prior runs.
	rawFileInTarget := fmt.Sprintf("/proc/%s/root%s", pid, rawFileName)
	_ = os.Remove(rawFileName)
	_ = os.Remove(rawFileInTarget)

	cmd := memrayCommand(p.commander, job, pid, rawFileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	log.DebugLogLn(fmt.Sprintf("memray attach stdout (PID %s): %q", pid, out.String()))
	log.DebugLogLn(fmt.Sprintf("memray attach stderr (PID %s): %q", pid, stderr.String()))
	if err != nil {
		stderrStr := stderr.String()
		// The PID may have exited between discovery and attach — skip it rather than
		// failing the entire profiling run.
		if strings.Contains(stderrStr, "No such file or directory") || strings.Contains(stderrStr, "No such process") {
			log.WarningLogLn(fmt.Sprintf("PID %s no longer exists, skipping: %s", pid, strings.TrimSpace(stderrStr)))
			return nil, time.Since(start)
		}
		// memray returns "An unexpected error occurred" when gdb injection fails (e.g. the process
		// was already injected in a previous session and can't be re-attached). Skip rather than
		// aborting all other PIDs; verbose output above will show the real gdb reason.
		if strings.Contains(stderrStr, "An unexpected error occurred") {
			log.WarningLogLn(fmt.Sprintf("PID %s could not be attached (already injected or gdb failure), skipping: %s", pid, strings.TrimSpace(stderrStr)))
			return nil, time.Since(start)
		}
		return errors.Wrapf(err, "could not launch profiler: %s", stderrStr), time.Since(start)
	}

	// memray attach exits immediately after injecting the tracker thread into the target process.
	// The tracker runs independently inside the target for job.Interval duration and writes the
	// output file to the target's filesystem. Wait for it to finish before reading the file.
	log.DebugLogLn(fmt.Sprintf("tracker injected for PID %s; waiting %v for profiling to complete", pid, job.Interval))
	time.Sleep(job.Interval)

	// The raw file was written by the target's native memray tracker into the target's /tmp.
	// Run the report command inside the target's mount namespace so the same memray version
	// is used for both writing and reading — avoiding format incompatibility errors.
	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)
	log.DebugLogLn(fmt.Sprintf("generating report for PID %s via target mount namespace", pid))
	err = handleReportInMountNs(p.commander, job, pid, rawFileName, resultFileName)
	if err != nil {
		log.ErrorLogLn(fmt.Sprintf("could not generate report (PID: %s): %s", pid, err.Error()))
		_ = os.Remove(rawFileInTarget)
		return err, time.Since(start)
	}
	_ = os.Remove(rawFileInTarget)

	return p.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

// handleReportInMountNs runs the memray report command inside the target's mount namespace
// (via nsenter --mount) so the target's own Python/memray installation is used. This avoids
// format incompatibility when the target has a different memray version than the profiling container.
// Exposed as a package-level var so tests can override it without /proc filesystem access.
var handleReportInMountNs = func(commander executil.Commander, job *job.ProfilingJob, pid string, rawFileName string, resultFileName string) error {
	mntNs := fmt.Sprintf("/proc/%s/ns/mnt", pid)

	var out bytes.Buffer
	var stderr bytes.Buffer

	pidNs := fmt.Sprintf("/proc/%s/ns/pid", pid)

	switch job.OutputType {
	case api.FlameGraph:
		// Output file goes into target's /tmp; we copy it back afterward.
		targetOutput := rawFileName + ".html"
		// Enter both mount and PID namespaces: mount so python3 uses the target's memray
		// installation, PID so /proc/self resolves correctly within the target's procfs.
		args := []string{"--mount=" + mntNs, "--pid=" + pidNs, "--", "python3", "-m", "memray", "flamegraph", rawFileName, "-o", targetOutput}
		cmd := commander.Command("nsenter", args...)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		log.DebugLogLn(fmt.Sprintf("nsenter --mount=%s --pid=%s -- python3 -m memray flamegraph %s -o %s", mntNs, pidNs, rawFileName, targetOutput))
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "could not generate flamegraph in target namespace: %s", stderr.String())
		}
		// Copy the output file from the target's /tmp back to the profiling container.
		targetOutputInHost := fmt.Sprintf("/proc/%s/root%s", pid, targetOutput)
		data, err := os.ReadFile(targetOutputInHost)
		if err != nil {
			return errors.Wrapf(err, "could not read flamegraph output from target namespace at %s", targetOutputInHost)
		}
		_ = os.Remove(targetOutputInHost)
		return os.WriteFile(resultFileName, data, 0600)

	case api.Summary, api.Tree:
		args := []string{"--mount=" + mntNs, "--pid=" + pidNs, "--", "python3", "-m", "memray", string(job.OutputType), rawFileName}
		cmd := commander.Command("nsenter", args...)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		log.DebugLogLn(fmt.Sprintf("nsenter --mount=%s --pid=%s -- python3 -m memray %s %s", mntNs, pidNs, string(job.OutputType), rawFileName))
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "could not generate %s report in target namespace: %s", string(job.OutputType), stderr.String())
		}
		return os.WriteFile(resultFileName, out.Bytes(), 0600)

	default:
		return fmt.Errorf("unsupported output type for memray: %s", job.OutputType)
	}
}

func (p *memrayManager) handleReport(job *job.ProfilingJob, rawFileName string, resultFileName string) error {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := memrayReportCommand(p.commander, job, rawFileName, resultFileName)
	if cmd == nil {
		return fmt.Errorf("unsupported output type for memray: %s", job.OutputType)
	}
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "could not generate %s report: %s", string(job.OutputType), stderr.String())
	}

	if job.OutputType != api.FlameGraph {
		// for summary/tree the report writes to stdout; persist it to the result file
		if err := os.WriteFile(resultFileName, []byte(out.String()), 0600); err != nil {
			return errors.Wrapf(err, "could not write %s report to file: %s", string(job.OutputType), resultFileName)
		}
	}

	return nil
}

func (p *MemrayProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
