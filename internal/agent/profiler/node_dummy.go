package profiler

import (
	"fmt"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
)

const (
	nodeDummySnapshotRetries = 120 // Two minutes
)

type NodeDummyProfiler struct {
	cwd string
	NodeDummyManager
}

type NodeDummyManager interface {
	invoke(*job.ProfilingJob, string, string) (error, time.Duration)
}

type nodeDummyManager struct {
	publisher       publish.Publisher
	snapshotRetries int
}

func NewNodeDummyProfiler(publisher publish.Publisher) *NodeDummyProfiler {
	return &NodeDummyProfiler{
		NodeDummyManager: &nodeDummyManager{
			publisher:       publisher,
			snapshotRetries: nodeDummySnapshotRetries,
		},
	}
}

func (n *NodeDummyProfiler) SetUp(job *job.ProfilingJob) error {
	targetFs, err := util.ContainerFileSystem(job.ContainerRuntime, job.ContainerID, job.ContainerRuntimePath)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The target filesystem is: %s", targetFs))

	cwd, err := util.GetCWD(job)
	if err != nil {
		return err
	}
	n.cwd = filepath.Join(targetFs, cwd)

	// remove previous files from a previous profiling
	file.RemoveAll(cwd, "*.heapsnapshot*")

	return nil
}

func (n *NodeDummyProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	rootPID, err := util.GetRootPID(job)
	if err != nil {
		return err, time.Since(start)
	}

	return n.invoke(job, util.GetFirstCandidatePID(rootPID), n.cwd)
}

var kill = func(pid, sig int) error {
	return syscall.Kill(pid, syscall.Signal(sig))
}

func (n *nodeDummyManager) invoke(job *job.ProfilingJob, pid, cwd string) (error, time.Duration) {
	start := time.Now()

	p, _ := strconv.Atoi(pid)
	err := kill(p, job.NodeHeapSnapshotSignal)
	if err != nil {
		return errors.Wrapf(err, "unable to send signal: %d to target process (PID: %s)", job.NodeHeapSnapshotSignal, pid), time.Since(start)
	}

	var retry int
	var fileName string
	fileSize := int64(1)
	var currentFileSize int64
	for range time.Tick(1 * time.Second) {
		fileName = file.First(filepath.Join(cwd, "*.heapsnapshot*"))
		if fileName != "" {
			currentFileSize = file.Size(fileName)
			if currentFileSize == fileSize {
				break
			}
		}
		if retry == n.snapshotRetries {
			return errors.Errorf("no heapsnapshot files found (PID: %s)", pid), time.Since(start)
		}
		fileSize = currentFileSize
		retry++
		log.DebugLogLn(fmt.Sprintf("No heapsnapshot available yet (PID: %s). Retrying...", pid))
	}

	// out file names is composed by the job info and the pid
	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, pid, job.Iteration)

	// copy the file to the tmp dir
	_, err = file.Copy(fileName, resultFileName)
	if err != nil {
		return errors.Wrapf(err, "could not move file: %s", fileName), time.Since(start)
	}

	// remove the file from the cwd
	err = file.Remove(fileName)
	if err != nil {
		log.WarningLogLn(fmt.Sprintf("The file could not be removed: %s", fileName))
	}

	return n.publisher.DoWithNativeGzipAndSplit(resultFileName, job.HeapDumpSplitInChunkSize, job.OutputType), time.Since(start)
}

func (n *NodeDummyProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)

	return nil
}
