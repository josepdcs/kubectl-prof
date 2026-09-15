//go:build integration

package jvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type target struct {
	pid string
	uid int
	// root bypasses the target's mount namespace, matching ContainerFileSystem.
	root string
	// procRoot honors mounts and read-only state in the target's namespace.
	procRoot string
}

func requireTarget(t *testing.T) target {
	t.Helper()

	pid := os.Getenv("TARGET_PID")
	if pid == "" {
		t.Skip("TARGET_PID is not set; run this suite through test/integration/jvm/run.sh")
	}
	uid, err := strconv.Atoi(os.Getenv("TARGET_UID"))
	require.NoError(t, err, "TARGET_UID has to be set alongside TARGET_PID")
	root := os.Getenv("TARGET_ROOTFS")
	require.NotEmpty(t, root, "TARGET_ROOTFS has to be set alongside TARGET_PID")

	return target{pid: pid, uid: uid, root: root, procRoot: filepath.Join("/proc", pid, "root")}
}

func runAsTargetUID(uid int, argv ...string) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(uid)},
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

// The bind-mounted staging directory must outlive every test that uses its inode.
var stageOnce sync.Once

func newTestManager(t *testing.T, tgt target) *asyncProfilerManager {
	t.Helper()

	publisher := publish.NewFakePublisher()
	publisher.On("Do").Return(nil)
	profiler := NewAsyncProfiler(executil.NewCommander(), publisher)
	manager := profiler.AsyncProfilerManager.(*asyncProfilerManager)

	var err error
	stageOnce.Do(func() { err = manager.stageProfilerLibrary(tgt.root) })
	require.NoError(t, err)

	manager.mutex.Lock()
	manager.libDir = filepath.Join(tgt.root, sharedDir)
	manager.mutex.Unlock()

	return manager
}

func TestIntegration_readTargetCreds(t *testing.T) {
	tgt := requireTarget(t)

	creds, err := readTargetCreds(tgt.pid)

	require.NoError(t, err)
	assert.Equal(t, tgt.uid, creds.uid, "the filesystem uid of the profiled JVM")
	assert.NotZero(t, creds.uid, "the whole defect depends on the target not being root")
}

func TestIntegration_stageProfilerLibrary(t *testing.T) {
	tgt := requireTarget(t)
	newTestManager(t, tgt)

	stagedLib := filepath.Join(tgt.root, sharedDir, libName)
	info, err := os.Stat(stagedLib)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))

	fromTarget := filepath.Join(tgt.procRoot, sharedDir, libName)
	assert.NoError(t, runAsTargetUID(tgt.uid, "cat", fromTarget))
	assert.Error(t, runAsTargetUID(tgt.uid, "sh", "-c", "echo tampered > "+fromTarget),
		"the workload must not be able to overwrite the staged library")

	_, err = os.Stat(filepath.Join(tgt.root, sharedDir, "build", "bin", "asprof"))
	assert.True(t, os.IsNotExist(err), "asprof must not be copied into the profiled container")
}

func TestIntegration_outputDirDeniesTheWorkload(t *testing.T) {
	tgt := requireTarget(t)

	dir, err := newOutputDir(tgt.pid)
	require.NoError(t, err)
	t.Cleanup(func() { _ = dir.remove() })

	info, err := os.Stat(dir.hostPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(outputDirMode), info.Mode().Perm())
	assert.Equal(t, uint32(0), info.Sys().(*syscall.Stat_t).Uid, "the result directory stays root-owned")

	f, err := dir.createResultFile("precreated.html", targetCreds{uid: tgt.uid, gid: tgt.uid})
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	hostFile := filepath.Join(dir.hostPath, "precreated.html")
	assert.NoError(t, runAsTargetUID(tgt.uid, "sh", "-c", "echo written > "+hostFile),
		"the JVM must be able to write the file handed to it")
	assert.Error(t, runAsTargetUID(tgt.uid, "sh", "-c", "echo nope > "+filepath.Join(dir.hostPath, "sibling.html")),
		"the JVM must not be able to create any other entry")
}

func TestIntegration_nonRootJVMProducesFlamegraph(t *testing.T) {
	tgt := requireTarget(t)
	manager := newTestManager(t, tgt)

	profilingJob := &job.ProfilingJob{
		Interval:   5 * time.Second,
		Event:      api.Ctimer,
		OutputType: api.FlameGraph,
		Tool:       api.AsyncProfiler,
		Compressor: compressor.None,
	}

	err, _ := manager.invoke(profilingJob, tgt.pid)
	require.NoError(t, err)

	result := localResult(profilingJob, tgt.pid)
	t.Cleanup(func() { _ = os.Remove(result) })

	content, err := os.ReadFile(result)
	require.NoError(t, err)
	assert.Greater(t, len(content), 1024, "expected a populated flamegraph")
	assert.Contains(t, string(content[:64]), "DOCTYPE html")
}

func TestIntegration_nonRootJVMProducesJFR(t *testing.T) {
	tgt := requireTarget(t)
	manager := newTestManager(t, tgt)

	profilingJob := &job.ProfilingJob{
		Interval:   5 * time.Second,
		Event:      api.Ctimer,
		OutputType: api.Jfr,
		Tool:       api.AsyncProfiler,
		Compressor: compressor.None,
	}

	err, _ := manager.invoke(profilingJob, tgt.pid)
	require.NoError(t, err)

	result := localResult(profilingJob, tgt.pid)
	t.Cleanup(func() { _ = os.Remove(result) })

	content, err := os.ReadFile(result)
	require.NoError(t, err)
	assert.Greater(t, len(content), 1024, "expected a populated recording")
	assert.Equal(t, "FLR\x00", string(content[:4]), "expected a JFR chunk header")
}

func TestIntegration_withoutPreCreationTheDumpFails(t *testing.T) {
	tgt := requireTarget(t)
	newTestManager(t, tgt)

	dir, err := newOutputDir(tgt.pid)
	require.NoError(t, err)
	t.Cleanup(func() { _ = dir.remove() })

	out, err := exec.Command(asprofBin,
		"--libpath", libContainerPath,
		"-o", "flamegraph",
		"-d", "2",
		"-e", "ctimer",
		"-f", dir.containerPath("never-created.html"),
		tgt.pid).CombinedOutput()

	require.Error(t, err)
	assert.Contains(t, string(out), "Could not open output file")
}

func localResult(profilingJob *job.ProfilingJob, pid string) string {
	return filepath.Join(common.TmpDir(),
		common.GetResultFile("", profilingJob.Tool, profilingJob.OutputType, pid, profilingJob.Iteration))
}
