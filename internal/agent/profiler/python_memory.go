package profiler

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	memrayStagingDir       = "/tmp/.kubectl-prof-memray"
	procRootPath           = "/proc/%s/root%s"
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

// libpythonRe matches lines in /proc/<pid>/maps that reference a libpython3.XX shared library.
var libpythonRe = regexp.MustCompile(`libpython(3\.\d+)`)

// pythonBinaryVersionRe extracts a version like "3.11" from a Python binary name (e.g. python3.11).
var pythonBinaryVersionRe = regexp.MustCompile(`python(3\.\d+)`)

// detectTargetPythonVersion determines the Python minor version of the target process.
// Primary: reads /proc/<pid>/maps to find the loaded libpython3.XX shared library.
// Fallback: resolves /proc/<pid>/exe symlink and checks if the binary name contains a version.
// Returns a string like "3.10", "3.11", "3.12", "3.13".
// Exposed as a package-level var so tests can override it.
var detectTargetPythonVersion = func(pid string) (string, error) {
	// Primary: read /proc/<pid>/maps for libpython reference.
	mapsPath := fmt.Sprintf("/proc/%s/maps", pid)
	mapsData, err := os.ReadFile(mapsPath)
	if err == nil {
		if m := libpythonRe.FindSubmatch(mapsData); m != nil {
			return string(m[1]), nil
		}
	}

	// Fallback: resolve /proc/<pid>/exe and extract version from binary name.
	exePath := fmt.Sprintf("/proc/%s/exe", pid)
	target, err := os.Readlink(exePath)
	if err == nil {
		base := filepath.Base(target)
		if m := pythonBinaryVersionRe.FindStringSubmatch(base); m != nil {
			return m[1], nil
		}
	}

	return "", fmt.Errorf("could not detect Python version for PID %s: no libpython in /proc/%s/maps and binary name has no version (statically-linked Python not supported)", pid, pid)
}

// copyDir delegates to file.CopyDir so it can be overridden in tests.
var copyDir = file.CopyDir

// stageMemrayPackage stages the version-matched memray package and _inject.abi3.so into the
// target container's filesystem. The full package is copied to /tmp/.kubectl-prof-memray/ so
// the patched payload's sys.path.insert can find it, and _inject.abi3.so is placed at its
// agent-install path so gdb's dlopen() can find it.
//
// Multiple PIDs from the same container share a mount namespace, so their
// /proc/<pid>/root/ paths all resolve to the same physical filesystem. The staging
// state is ref-counted: the first caller stages the files; subsequent callers increment
// the counter. The last cleanup decrements to zero and performs the actual removal,
// trying each known PID's /proc path in turn so that a stale path (from a PID that
// exited mid-profiling) does not prevent cleanup.
var (
	stageMemrayPackageMu        sync.Mutex
	stageMemrayPackageRefs      = map[string]int{}
	stageMemrayPackageOriginals = map[string][]byte{}
	stageMemrayPackageHadOrig   = map[string]bool{}
	stageMemrayPackagePIDs      = map[string][]string{}
)

var stageMemrayPackage = func(pid string) (func(), error) {
	// Detect target Python version to select the correct memray build.
	pyVersion, err := detectTargetPythonVersion(pid)
	if err != nil {
		return nil, fmt.Errorf("could not detect target Python version for memray staging: %w", err)
	}

	// Locate _inject.abi3.so in the agent's memray installation.
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
	libDst := fmt.Sprintf(procRootPath, pid, libSrc)

	// Version-matched memray package source and target staging directory.
	pkgSrc := fmt.Sprintf("/opt/memray/%s", pyVersion)
	pkgDst := fmt.Sprintf(procRootPath, pid, memrayStagingDir)

	if err := os.MkdirAll(filepath.Dir(libDst), 0755); err != nil {
		return nil, fmt.Errorf("could not create directory for %s: %w", libDst, err)
	}

	libData, err := os.ReadFile(libSrc)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", libSrc, err)
	}

	stageMemrayPackageMu.Lock()
	if stageMemrayPackageRefs[libSrc] == 0 {
		// First reference: stage inject library and memray package.
		oldData, _ := os.ReadFile(libDst)
		stageMemrayPackageHadOrig[libSrc] = oldData != nil
		stageMemrayPackageOriginals[libSrc] = oldData
		if err := os.WriteFile(libDst, libData, 0755); err != nil {
			stageMemrayPackageMu.Unlock()
			return nil, fmt.Errorf("could not stage %s to %s: %w", libSrc, libDst, err)
		}
		// Stage full memray package for import.
		if err := copyDir(pkgSrc, pkgDst); err != nil {
			// Rollback: remove the already-staged inject library.
			if stageMemrayPackageHadOrig[libSrc] {
				_ = os.WriteFile(libDst, stageMemrayPackageOriginals[libSrc], 0755)
			} else {
				_ = os.Remove(libDst)
			}
			stageMemrayPackageMu.Unlock()
			return nil, fmt.Errorf("could not stage memray package from %s to %s: %w", pkgSrc, pkgDst, err)
		}
	}
	stageMemrayPackageRefs[libSrc]++
	stageMemrayPackagePIDs[libSrc] = append(stageMemrayPackagePIDs[libSrc], pid)
	stageMemrayPackageMu.Unlock()

	return func() {
		stageMemrayPackageMu.Lock()
		defer stageMemrayPackageMu.Unlock()
		stageMemrayPackageRefs[libSrc]--
		if stageMemrayPackageRefs[libSrc] == 0 {
			// Try each known PID's /proc path until one succeeds.
			restored := false
			for _, p := range stageMemrayPackagePIDs[libSrc] {
				// Remove staged memray package directory.
				stagingPath := fmt.Sprintf(procRootPath, p, memrayStagingDir)
				_ = os.RemoveAll(stagingPath)

				// Restore or remove _inject.abi3.so.
				dst := fmt.Sprintf(procRootPath, p, libSrc)
				var restoreErr error
				if stageMemrayPackageHadOrig[libSrc] {
					restoreErr = os.WriteFile(dst, stageMemrayPackageOriginals[libSrc], 0755)
				} else {
					restoreErr = os.Remove(dst)
				}
				if restoreErr == nil {
					restored = true
					break
				}
			}
			if !restored {
				log.WarningLogLn(fmt.Sprintf("could not restore memray inject library %s: all PID paths exhausted (library may remain staged)", libSrc))
			}
			delete(stageMemrayPackageRefs, libSrc)
			delete(stageMemrayPackageOriginals, libSrc)
			delete(stageMemrayPackageHadOrig, libSrc)
			delete(stageMemrayPackagePIDs, libSrc)
		}
	}, nil
}

// pidExists checks whether a process with the given PID still exists by stat-ing /proc/<pid>.
// Returns false only when the error is definitively "not exist"; any other error (EACCES,
// transient procfs failure) returns true so that callers do not incorrectly skip a live PID.
// Exposed as a package-level var so tests can override it.
var pidExists = func(pid string) bool {
	_, err := os.Stat(fmt.Sprintf("/proc/%s", pid))
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

// rawFileInTargetPath returns the path to the raw file inside the target's mount namespace.
// Exposed as a package-level var so tests can override it.
var rawFileInTargetPath = func(pid string, rawFileName string) string {
	return fmt.Sprintf(procRootPath, pid, rawFileName)
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

// waitForProfilingInterval waits for the profiling interval to elapse, emitting periodic
// heartbeat events to keep the log stream alive through proxies/load balancers.
// If heartbeatInterval is zero or negative the default of 30 seconds is used.
func waitForProfilingInterval(interval, heartbeatInterval time.Duration) {
	if heartbeatInterval <= 0 {
		heartbeatInterval = 30 * time.Second
	}
	remaining := interval
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
}

func (p *memrayManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	var out bytes.Buffer
	var stderr bytes.Buffer

	// Stage the memray package and _inject.abi3.so into the target container's filesystem.
	// Staging is required for profiling — without it, gdb injection and import memray will fail.
	cleanup, err := stageMemrayPackage(pid)
	if err != nil {
		// The PID may have exited between discovery and staging — if the process is gone,
		// skip it rather than failing the entire profiling run.
		if !pidExists(pid) {
			log.WarningLogLn(fmt.Sprintf("PID %s no longer exists during staging, skipping: %s", pid, err))
			return nil, time.Since(start)
		}
		return fmt.Errorf("could not stage memray package for PID %s: %w", pid, err), time.Since(start)
	}
	defer cleanup()

	// intermediate raw binary file (not a user-facing output type, so built directly)
	rawFileName := filepath.Join(common.TmpDir(), fmt.Sprintf("%smemray-raw-%s-%d.bin", config.ProfilingPrefix, pid, job.Iteration))

	// The tracker (--aggregate) writes the raw file to its own /tmp (target's mount namespace).
	// Clean both locations before running to avoid "file exists" errors from prior runs.
	rawFileInTarget := rawFileInTargetPath(pid, rawFileName)
	_ = file.Remove(rawFileName)
	_ = file.Remove(rawFileInTarget)

	cmd := memrayCommand(p.commander, job, pid, rawFileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
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
	waitForProfilingInterval(job.Interval, job.HeartbeatInterval)

	// Verify the tracker actually produced a raw file before attempting report generation.
	// The tracker may have crashed or the target process may have exited mid-profile.
	if _, err := checkRawFileExists(rawFileInTarget); err != nil {
		log.WarningLogLn(fmt.Sprintf("raw file not found for PID %s at %s after profiling interval (target may have exited): %s", pid, rawFileInTarget, err))
		return nil, time.Since(start)
	}

	// The raw file was written by the target's native memray tracker into the target's /tmp.
	// Copy it to the agent's local /tmp and run the report command locally using the agent's
	// own memray installation — no need to enter the target's mount namespace.
	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)
	log.DebugLogLn(fmt.Sprintf("generating report for PID %s locally in agent container", pid))
	err = handleReport(p.commander, job, pid, rawFileName, resultFileName)
	if err != nil {
		log.ErrorLogLn(fmt.Sprintf("could not generate report (PID: %s): %s", pid, err.Error()))
		_ = os.Remove(rawFileInTarget)
		return err, time.Since(start)
	}
	_ = file.Remove(rawFileInTarget)

	return p.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

// copyRawFileFromTarget copies the raw profiling data from the target's filesystem
// (via /proc/<pid>/root/...) to a local path in the agent container.
// Exposed as a package-level var so tests can override it.
var copyRawFileFromTarget = func(src, dst string) error {
	_, err := file.Copy(src, dst)
	return err
}

// handleReport copies the raw profiling data from the target's filesystem to the agent's
// local /tmp, then runs the memray report command locally using the agent's own memray
// installation. This avoids entering the target's mount namespace for report generation.
// Exposed as a package-level var so tests can override it.
var handleReport = func(commander executil.Commander, job *job.ProfilingJob, pid string, rawFileName string, resultFileName string) error {
	// Copy raw file from target's filesystem to agent's local /tmp.
	rawFileInTarget := rawFileInTargetPath(pid, rawFileName)
	localRawFile := rawFileName + ".local"
	if err := copyRawFileFromTarget(rawFileInTarget, localRawFile); err != nil {
		return err
	}
	defer func(f string) {
		err := file.Remove(f)
		if err != nil {
			log.WarningLogLn(fmt.Sprintf("could not remove local raw file %s: %s", f, err))
		}
	}(localRawFile)

	var out bytes.Buffer
	var stderr bytes.Buffer

	switch job.OutputType {
	case api.FlameGraph:
		localOutput := localRawFile + ".html"
		args := []string{"-m", "memray", "flamegraph", localRawFile, "-o", localOutput}
		cmd := commander.Command("python3", args...)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "could not generate flamegraph: %s", stderr.String())
		}
		if _, err := file.Copy(localOutput, resultFileName); err != nil {
			return errors.Wrapf(err, "could not copy flamegraph output to %s", resultFileName)
		}
		_ = file.Remove(localOutput)
		return nil

	case api.Summary:
		args := []string{"-m", "memray", string(job.OutputType), localRawFile}
		cmd := commander.Command("python3", args...)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "could not generate %s report: %s", string(job.OutputType), stderr.String())
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
