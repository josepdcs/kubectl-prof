package jvm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
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
	// sharedDir must remain root-owned because the agent stages executable code here.
	sharedDir = "/kubectl-prof"

	// The JVM resolves libName in its own mount namespace.
	libName = "libasyncProfiler.so"

	// asprof runs from the agent image to prevent the workload from replacing it.
	libSourcePath = "/app/async-profiler/build/lib/" + libName
	asprofBin     = "/app/async-profiler/build/bin/asprof"

	libFileMode = 0o644
	libDirMode  = 0o755

	asyncProfilerDelayBetweenJobs = 2 * time.Second
)

var libContainerPath = filepath.Join(sharedDir, libName)

var asyncProfilerCommand = func(j *asyncProfilerManager, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	event := string(job.Event)
	output := string(job.OutputType)
	if job.OutputType == api.Raw {
		// overrides to collapsed type since it is the type defined be async-profiler, which it is what we want
		output = string(api.Collapsed)
	}
	args := []string{
		"--libpath", libContainerPath,
		"-o", output,
		"-d", interval,
		"-f", fileName,
		"-e", event,
		"--fdtransfer",
	}

	// Add additional async-profiler arguments if they exist
	if job.AdditionalArguments != nil {
		// Iterate through additional arguments in order
		i := 0
		for {
			key := fmt.Sprintf("async-profiler-arg-%d", i)
			if arg, ok := job.AdditionalArguments[key]; ok {
				args = append(args, arg)
				i++
			} else {
				break
			}
		}
	}

	args = append(args, pid)

	return j.commander.Command(asprofBin, args...)
}

var asyncProfilerStopCommand = func(j *asyncProfilerManager, job *job.ProfilingJob, pid string) *exec.Cmd {
	return j.commander.Command(
		asprofBin,
		"stop",
		"--libpath", libContainerPath,
		pid)
}

type AsyncProfiler struct {
	targetPIDs []string
	delay      time.Duration
	AsyncProfilerManager
}

type AsyncProfilerManager interface {
	stageProfilerLibrary(targetFs string) error
	invoke(*job.ProfilingJob, string) (error, time.Duration)
	cleanUp(*job.ProfilingJob, string)
	releaseTargetArtifacts()
}

type asyncProfilerManager struct {
	commander executil.Commander
	publisher publish.Publisher

	mutex      sync.Mutex
	outputDirs map[string]*outputDir
	libDir     string
}

func NewAsyncProfiler(commander executil.Commander, publisher publish.Publisher) *AsyncProfiler {
	return &AsyncProfiler{
		delay: asyncProfilerDelayBetweenJobs,
		AsyncProfilerManager: &asyncProfilerManager{
			commander:  commander,
			publisher:  publisher,
			outputDirs: make(map[string]*outputDir),
		},
	}
}

func (j *AsyncProfiler) SetUp(job *job.ProfilingJob) error {
	targetFs, err := util.ContainerFileSystem(job.ContainerRuntime, job.ContainerID, job.ContainerRuntimePath)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The target filesystem is: %s", targetFs))

	if stringUtils.IsNotBlank(job.PID) {
		j.targetPIDs = []string{job.PID}
	} else {
		pids, err := util.GetCandidatePIDs(job)
		if err != nil {
			return err
		}
		log.DebugLogLn(fmt.Sprintf("The PIDs to be profiled: %s", pids))
		j.targetPIDs = pids
	}

	return j.stageProfilerLibrary(targetFs)
}

// stageProfilerLibrary copies the library through the host-side container filesystem.
// This path remains writable when the container's root filesystem is read-only in its
// own mount namespace. The root-owned directory prevents the workload from replacing
// the library; asprof remains in the agent image for the same reason.
func (j *asyncProfilerManager) stageProfilerLibrary(targetFs string) error {
	libDir := filepath.Join(targetFs, sharedDir)
	log.DebugLogLn(fmt.Sprintf("Staging %s into '%s'", libName, libDir))

	if err := safeMkdir(libDir); err != nil {
		return err
	}

	source, err := os.Open(libSourcePath)
	if err != nil {
		return errors.Wrapf(err, "could not open %s", libSourcePath)
	}
	defer func() { _ = source.Close() }()

	targetPath := filepath.Join(libDir, libName)
	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY|syscall.O_NOFOLLOW, libFileMode)
	if err != nil {
		return errors.Wrapf(err, "could not stage %s into the target container", libName)
	}

	if _, err := io.Copy(target, source); err != nil {
		_ = target.Close()
		_ = os.Remove(targetPath)
		return errors.Wrapf(err, "could not copy %s into the target container", libName)
	}
	if err := target.Close(); err != nil {
		_ = os.Remove(targetPath)
		return errors.Wrapf(err, "could not close the staged %s", libName)
	}

	j.mutex.Lock()
	j.libDir = libDir
	j.mutex.Unlock()

	return nil
}

// safeMkdir refuses symlinks so the workload cannot redirect the staging writes.
func safeMkdir(dir string) error {
	info, err := os.Lstat(dir)
	switch {
	case err == nil && info.Mode()&os.ModeSymlink != 0:
		return errors.Errorf("refusing to use %s: it is a symlink", dir)
	case err == nil && !info.IsDir():
		return errors.Errorf("refusing to use %s: it is not a directory", dir)
	case err == nil:
		return validatePrivateDir(dir, info)
	case !os.IsNotExist(err):
		return errors.Wrapf(err, "could not inspect %s", dir)
	}

	if err := os.Mkdir(dir, libDirMode); err != nil {
		return errors.Wrapf(err, "could not create the shared directory %s in the target container", dir)
	}
	info, err = os.Lstat(dir)
	if err != nil {
		return errors.Wrapf(err, "could not inspect the shared directory %s", dir)
	}
	return validatePrivateDir(dir, info)
}

func validatePrivateDir(dir string, info os.FileInfo) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.Errorf("could not determine the owner of %s", dir)
	}
	if int(stat.Uid) != os.Geteuid() {
		return errors.Errorf("refusing to use %s: it is not owned by uid %d", dir, os.Geteuid())
	}
	if info.Mode().Perm()&0o022 != 0 {
		return errors.Errorf("refusing to use %s: it is writable by group or others", dir)
	}
	return nil
}

func (j *AsyncProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	pool := pond.New(len(j.targetPIDs), 0, pond.MinWorkers(len(j.targetPIDs)))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to profile
	for _, pid := range j.targetPIDs {
		pid := pid
		group.Submit(func() error {
			err, _ := j.invoke(job, pid)
			return err
		})
		// wait a bit between jobs for not overloading the system
		time.Sleep(j.delay)
	}

	// wait for all tasks to finish
	err := group.Wait()

	return err, time.Since(start)
}

func (j *asyncProfilerManager) outputDirFor(pid string) (*outputDir, error) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if dir, ok := j.outputDirs[pid]; ok {
		return dir, nil
	}

	dir, err := newOutputDir(pid)
	if err != nil {
		return nil, err
	}
	j.outputDirs[pid] = dir

	return dir, nil
}

func (j *asyncProfilerManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()
	var out bytes.Buffer
	var stderr bytes.Buffer

	creds, err := readTargetCredsFn(pid)
	if err != nil {
		return err, time.Since(start)
	}

	dir, err := j.outputDirFor(pid)
	if err != nil {
		return err, time.Since(start)
	}

	// Pre-creation lets the workload write the file without making its parent writable.
	resultFileName := common.GetResultFile("", job.Tool, job.OutputType, pid, job.Iteration)
	resultFile, err := dir.createResultFile(resultFileName, creds)
	if err != nil {
		return err, time.Since(start)
	}
	defer func() { _ = resultFile.Close() }()

	cmd := asyncProfilerCommand(j, job, pid, dir.containerPath(resultFileName))
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}
	log.DebugLogLn(out.String())

	// Reuse the descriptor so a workload cannot redirect the privileged read.
	localFile := filepath.Join(common.TmpDir(), resultFileName)
	if err := copyResult(resultFile, localFile); err != nil {
		return err, time.Since(start)
	}

	return j.publisher.Do(job.Compressor, localFile, job.OutputType), time.Since(start)
}

func copyResult(src *os.File, dst string) error {
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return errors.Wrap(err, "could not rewind the result file")
	}

	local, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return errors.Wrapf(err, "could not create the local result file %s", dst)
	}
	if _, err := io.Copy(local, src); err != nil {
		_ = local.Close()
		_ = os.Remove(dst)
		return errors.Wrapf(err, "could not copy the result into %s", dst)
	}
	if err := local.Close(); err != nil {
		_ = os.Remove(dst)
		return errors.Wrapf(err, "could not close the local result file %s", dst)
	}

	return nil
}

func (j *AsyncProfiler) CleanUp(job *job.ProfilingJob) error {
	for _, pid := range j.targetPIDs {
		j.cleanUp(job, pid)
	}

	j.releaseTargetArtifacts()
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix+string(job.OutputType))

	return nil
}

func (j *asyncProfilerManager) releaseTargetArtifacts() {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	for pid, dir := range j.outputDirs {
		if err := dir.remove(); err != nil {
			log.WarningLogLn(fmt.Sprintf("result directory for PID %s could not be removed: %s", pid, err))
		}
		delete(j.outputDirs, pid)
	}

	if j.libDir != "" {
		if err := os.RemoveAll(j.libDir); err != nil {
			log.WarningLogLn(fmt.Sprintf("async-profiler library could not be removed: %s", err))
		}
		j.libDir = ""
	}
}

func (j *asyncProfilerManager) cleanUp(job *job.ProfilingJob, pid string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := asyncProfilerStopCommand(j, job, pid)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.WarningLogLn(stderr.String())
	}
	_, _ = fmt.Fprint(io.Discard, out.String())
}
