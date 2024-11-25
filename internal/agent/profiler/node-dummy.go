package profiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/samber/lo"
)

const (
	nodeDummyDelayBetweenJobs = 5 * time.Second
	nodeDummySnapshotRetries  = 120 // Two minutes
)

type NodeDummyProfiler struct {
	delay time.Duration
	NodeDummyManager
}

type NodeDummyManager interface {
	removeTmpDir() error
	linkTmpDirToTargetTmpDir(string) error
	invoke(*job.ProfilingJob, string) (error, time.Duration)
}

type nodeDummyManager struct {
	publisher publish.Publisher
}

func NewNodeDummyProfiler(publisher publish.Publisher) *NodeDummyProfiler {
	return &NodeDummyProfiler{
		delay: nodeDummyDelayBetweenJobs,
		NodeDummyManager: &nodeDummyManager{
			publisher: publisher,
		},
	}
}

func (n *NodeDummyProfiler) SetUp(job *job.ProfilingJob) error {
	targetFs, err := util.ContainerFileSystem(job.ContainerRuntime, job.ContainerID, job.ContainerRuntimePath)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The target filesystem is: %s", targetFs))

	err = n.removeTmpDir()
	if err != nil {
		return err
	}

	targetTmpDir := filepath.Join(targetFs, "tmp")
	// remove previous files from a previous profiling
	file.RemoveAll(targetTmpDir, config.ProfilingPrefix+string(job.OutputType))

	return n.linkTmpDirToTargetTmpDir(targetTmpDir)
}

func (n *nodeDummyManager) removeTmpDir() error {
	return os.RemoveAll(common.TmpDir())
}

func (n *nodeDummyManager) linkTmpDirToTargetTmpDir(targetTmpDir string) error {
	return os.Symlink(targetTmpDir, common.TmpDir())
}

func (n *NodeDummyProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	rootPID, err := util.GetRootPID(job)
	if err != nil {
		return err, time.Since(start)
	}

	return n.invoke(job, rootPID)
}

func (n *nodeDummyManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()

	p, _ := strconv.Atoi(pid)
	err := syscall.Kill(p, syscall.Signal(job.NodeHeapSnapshotSignal))
	if err != nil {
		log.ErrorLogLn(fmt.Sprintf("unable to send signal: %d to target process (PID: %s): %s", job.NodeHeapSnapshotSignal, pid, err.Error()))
		return nil, time.Since(start)
	}

	var retry int
	var files []string
	for range time.Tick(1 * time.Second) {
		files = file.List(filepath.Join(common.TmpDir(), "*.heapsnapshot*"))
		if len(files) == 0 && retry == nodeDummySnapshotRetries {
			log.ErrorLogLn(fmt.Sprintf("no heapsnapshot files found (PID: %s)", pid))
			return nil, time.Since(start)
		}
		retry++
	}

	fileName, ok := lo.Find(files, func(f string) bool { return strings.Contains(f, pid) })
	if !ok {
		log.ErrorLogLn(fmt.Sprintf("no heapsnapshot files found (PID: %s)", pid))
		return nil, time.Since(start)
	}
	log.DebugLogLn(fmt.Sprintf("The heapsnapshot file is: %s", fileName))

	// out file names is composed by the job info and the pid
	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)

	err = os.Rename(fileName, resultFileName)
	if err != nil {
		log.ErrorLogLn(fmt.Sprintf("could not rename file (PID: %s): %s", pid, err.Error()))
		return nil, time.Since(start)
	}

	return n.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (n *NodeDummyProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
