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
	"sync"
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
// Only _inject.abi3.so is staged; the target's native _memray.cpython-*.so is intentionally
// left untouched so the tracker writes a file in the target's format (read back via nsenter).
// Returns a cleanup function that restores the original file (or removes the staged one).
//
// Multiple PIDs from the same container share a mount namespace, so their
// /proc/<pid>/root/<libSrc> paths all resolve to the same physical file. The staging
// state is ref-counted by libSrc: the first caller stages the file and saves the original;
// subsequent callers for the same libSrc increment the counter without re-reading. The
// last cleanup decrements to zero and performs the actual restore/remove, trying each known
// PID's /proc path in turn so that a stale path (from a PID that exited mid-profiling) does
// not prevent restoration.
var (
	stageMemrayLibMu        sync.Mutex
	stageMemrayLibRefs      = map[string]int{}
	stageMemrayLibOriginals = map[string][]byte{}
	stageMemrayLibHadOrig   = map[string]bool{}
	stageMemrayLibPIDs      = map[string][]string{}
)

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

	stageMemrayLibMu.Lock()
	if stageMemrayLibRefs[libSrc] == 0 {
		// First reference: save original and stage the file.
		oldData, _ := os.ReadFile(libDst)
		stageMemrayLibHadOrig[libSrc] = oldData != nil
		stageMemrayLibOriginals[libSrc] = oldData
		if err := os.WriteFile(libDst, data, 0755); err != nil {
			stageMemrayLibMu.Unlock()
			return nil, fmt.Errorf("could not stage %s to %s: %w", libSrc, libDst, err)
		}
	}
	stageMemrayLibRefs[libSrc]++
	stageMemrayLibPIDs[libSrc] = append(stageMemrayLibPIDs[libSrc], pid)
	stageMemrayLibMu.Unlock()

	return func() {
		stageMemrayLibMu.Lock()
		defer stageMemrayLibMu.Unlock()
		stageMemrayLibRefs[libSrc]--
		if stageMemrayLibRefs[libSrc] == 0 {
			// Try each known PID's /proc path until one succeeds. The target PID that
			// registered this cleanup may have exited by now, making its
			// /proc/<pid>/root/... path stale. All PIDs in the same container share a
			// mount namespace, so any still-live PID's path reaches the same physical
			// file. Iterating through the full set avoids a silent failure leaving the
			// staged library in place.
			restored := false
			for _, p := range stageMemrayLibPIDs[libSrc] {
				dst := fmt.Sprintf("/proc/%s/root%s", p, libSrc)
				var err error
				if stageMemrayLibHadOrig[libSrc] {
					err = os.WriteFile(dst, stageMemrayLibOriginals[libSrc], 0755)
				} else {
					err = os.Remove(dst)
				}
				if err == nil {
					restored = true
					break
				}
			}
			if !restored {
				log.WarningLogLn(fmt.Sprintf("could not restore memray inject library %s: all PID paths exhausted (library may remain staged)", libSrc))
			}
			delete(stageMemrayLibRefs, libSrc)
			delete(stageMemrayLibOriginals, libSrc)
			delete(stageMemrayLibHadOrig, libSrc)
			delete(stageMemrayLibPIDs, libSrc)
		}
	}, nil
}

// rawFileInTargetPath returns the path to the raw file inside the target's mount namespace.
// Exposed as a package-level var so tests can override it.
var rawFileInTargetPath = func(pid string, rawFileName string) string {
	return fmt.Sprintf("/proc/%s/root%s", pid, rawFileName)
}

// checkRawFileExists verifies the raw file exists. Exposed as a package-level var so tests
// can override it without requiring /proc filesystem access.
var checkRawFileExists = func(path string) (os.FileInfo, error) {
	return os.Stat(path)
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

	// intermediate raw binary file (not a user-facing output type, so built directly)
	rawFileName := filepath.Join(common.TmpDir(), fmt.Sprintf("%smemray-raw-%s-%d.bin", config.ProfilingPrefix, pid, job.Iteration))

	// The tracker (--aggregate) writes the raw file to its own /tmp (target's mount namespace).
	// Clean both locations before running to avoid "file exists" errors from prior runs.
	rawFileInTarget := rawFileInTargetPath(pid, rawFileName)
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
		// failing the entire profiling run. Restrict "No such file or directory" to
		// errors that reference the PID-specific /proc/<pid>/ path so that unrelated
		// failures (e.g. missing /app/memray binary) are not silently swallowed.
		if (strings.Contains(stderrStr, "No such file or directory") && strings.Contains(stderrStr, fmt.Sprintf("/proc/%s/", pid))) ||
			strings.Contains(stderrStr, "No such process") {
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
	// Emit periodic heartbeat events to keep the log stream alive through proxies/load balancers.
	log.DebugLogLn(fmt.Sprintf("tracker injected for PID %s; waiting %v for profiling to complete", pid, job.Interval))
	heartbeatInterval := 30 * time.Second
	remaining := job.Interval
	for remaining > 0 {
		sleep := heartbeatInterval
		if remaining < sleep {
			sleep = remaining
		}
		time.Sleep(sleep)
		remaining -= sleep
		if remaining > 0 {
			_ = log.EventLn(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Profiling})
		}
	}

	// Verify the tracker actually produced a raw file before attempting report generation.
	// The tracker may have crashed or the target process may have exited mid-profile.
	if _, err := checkRawFileExists(rawFileInTarget); err != nil {
		log.WarningLogLn(fmt.Sprintf("raw file not found for PID %s at %s after profiling interval (target may have exited): %s", pid, rawFileInTarget, err))
		return nil, time.Since(start)
	}

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
		if err := os.WriteFile(resultFileName, data, 0600); err != nil {
			return err
		}
		_ = os.Remove(targetOutputInHost)
		return nil

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

func (p *MemrayProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
