package jvm

import (
	"bytes"
	"context"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/alitto/pond"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const (
	jcmd                     = "/opt/jdk/bin/jcmd"
	jfrSettingsImageFilePath = "/app/jfr/settings/jfr-profile.jfc"
	jfrSettingsTmpFilePath   = "/tmp/jfr-profile.jfc"
	invocationName           = "name=pid_"
	jcmdDelayBetweenJobs     = 2 * time.Second
)

var silentJcmdCommander = executil.NewSilentCommander()

var recordingPIDs chan string

var jcmdCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	switch job.OutputType {
	case api.ThreadDump:
		args := []string{pid, "Thread.print"}
		return commander.Command(jcmd, args...)
	case api.HeapDump:
		args := []string{pid, "GC.heap_dump", fileName}
		return commander.Command(jcmd, args...)
	case api.HeapHistogram:
		args := []string{pid, "GC.class_histogram"}
		return commander.Command(jcmd, args...)
	}

	// default api.Jfr
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{pid, "JFR.start", "duration=" + interval + "s", "filename=" + fileName, invocationName + pid + "_" + string(job.OutputType), "settings=" + jfrSettingsTmpFilePath}
	return commander.Command(jcmd, args...)
}

var jcmdStopCommand = func(commander executil.Commander, job *job.ProfilingJob, pid string) *exec.Cmd {
	return commander.Command(jcmd, pid, "JFR.stop", invocationName+pid+"_"+string(job.OutputType))
}

type JcmdProfiler struct {
	targetPIDs []string
	delay      time.Duration
	JcmdManager
}

type JcmdManager interface {
	removeTmpDir() error
	linkTmpDirToTargetTmpDir(string) error
	copyJfrSettingsToTmpDir() error
	invoke(*job.ProfilingJob, string) (error, time.Duration)
	handleProfilingResult(job *job.ProfilingJob, fileName string, out bytes.Buffer, targetPID string) error
	handleJcmdRecording(targetPID string, outputType string)
	publishResult(compressor compressor.Type, fileName string, outputType api.OutputType, heapDumpSplitInChunkSize string) error
	cleanUp(*job.ProfilingJob, string)
}

type jcmdManager struct {
	commander executil.Commander
	publisher publish.Publisher
}

func NewJcmdProfiler(commander executil.Commander, publisher publish.Publisher) *JcmdProfiler {
	return &JcmdProfiler{
		delay: jcmdDelayBetweenJobs,
		JcmdManager: &jcmdManager{
			commander: commander,
			publisher: publisher,
		}}
}

func (j *JcmdProfiler) SetUp(job *job.ProfilingJob) error {
	targetFs, err := util.ContainerFileSystem(job.ContainerRuntime, job.ContainerID)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The target filesystem is: %s", targetFs))

	err = j.removeTmpDir()
	if err != nil {
		return err
	}

	targetTmpDir := filepath.Join(targetFs, "tmp")
	// remove previous files from a previous profiling
	file.RemoveAll(targetTmpDir, config.ProfilingPrefix+string(job.OutputType))

	err = j.linkTmpDirToTargetTmpDir(targetTmpDir)
	if err != nil {
		return err
	}

	if stringUtils.IsNotBlank(job.PID) {
		j.targetPIDs = []string{job.PID}
		recordingPIDs = make(chan string, 1)
	} else {
		pids, err := util.GetCandidatePIDs(job)
		if err != nil {
			return err
		}
		log.DebugLogLn(fmt.Sprintf("The PIDs to be profiled: %s", pids))
		j.targetPIDs = pids
		recordingPIDs = make(chan string, len(pids))
	}

	return j.copyJfrSettingsToTmpDir()
}

func (j *jcmdManager) removeTmpDir() error {
	return os.RemoveAll(common.TmpDir())
}

func (j *jcmdManager) linkTmpDirToTargetTmpDir(targetTmpDir string) error {
	return os.Symlink(targetTmpDir, common.TmpDir())
}

func (j *jcmdManager) copyJfrSettingsToTmpDir() error {
	cmd := executil.Command("cp", jfrSettingsImageFilePath, common.TmpDir())
	return cmd.Run()
}

func (j *JcmdProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
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

func (j *jcmdManager) invoke(job *job.ProfilingJob, pid string) (error, time.Duration) {
	start := time.Now()
	var out bytes.Buffer
	var stderr bytes.Buffer

	resultFileName := common.GetResultFileWithPID(common.TmpDir(), job.Tool, job.OutputType, pid)
	cmd := jcmdCommand(j.commander, job, pid, resultFileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return errors.Wrapf(err, "could not launch profiler: %s", stderr.String()), time.Since(start)
	}

	err = j.handleProfilingResult(job, resultFileName, out, pid)
	if err != nil {
		return err, time.Since(start)
	}

	return j.publishResult(job.Compressor, resultFileName, job.OutputType, job.HeapDumpSplitInChunkSize), time.Since(start)
}

func (j *jcmdManager) handleProfilingResult(job *job.ProfilingJob, fileName string, out bytes.Buffer, pid string) error {
	switch job.OutputType {
	case api.Jfr:
		j.handleJcmdRecording(pid, string(job.OutputType))
	case api.ThreadDump, api.HeapHistogram:
		err := os.WriteFile(fileName, out.Bytes(), 0644)
		if err != nil {
			return errors.Wrap(err, "could not save dump to file")
		}
	default:
		log.DebugLogLn(out.String())
	}

	return nil
}

func (j *jcmdManager) handleJcmdRecording(pid string, outputType string) {
	done := make(chan bool)

	go func(pid string) {
		log.DebugLogLn(fmt.Sprintf("Jcmd recording started for PID %s", pid))
		for {
			select {
			case currentPID := <-recordingPIDs:
				if currentPID == pid {
					log.WarningLogLn(fmt.Sprintf("Stopping is detected. Ignoring jcmd recording for PID %s", pid))
					return
				}
			default:
				var out bytes.Buffer
				var stderr bytes.Buffer
				cmd := silentJcmdCommander.Command(jcmd, pid, "JFR.check", "name=pid_"+pid+"_"+outputType)
				cmd.Stdout = &out
				cmd.Stderr = &stderr
				err := cmd.Run()
				if err != nil {
					log.ErrorLogLn(stderr.String())
					done <- true
					return
				}
				outputTxt := out.String()
				if !stringUtils.Contains(outputTxt, "running") && !stringUtils.Contains(outputTxt, "delayed") {
					printErrorIfStopped(outputTxt)
					done <- true
					return
				}

			}
			time.Sleep(1 * time.Second)
		}
	}(pid)

	<-done
}

func printErrorIfStopped(outputTxt string) {
	// fix: in case of any error in a previous profiling which made that jcmd remains stopped
	if stringUtils.Contains(outputTxt, "stopped") {
		log.ErrorLogLn("Jcmd remains stopped for unknown reason, please restart the pod to be profiled")
	}
}

func (j *jcmdManager) publishResult(c compressor.Type, fileName string, outputType api.OutputType, heapDumpSplitInChunkSize string) error {
	if outputType == api.HeapDump {
		return j.publisher.DoWithNativeGzipAndSplit(fileName, heapDumpSplitInChunkSize, outputType)
	}
	return j.publisher.Do(c, fileName, outputType)
}

func (j *JcmdProfiler) CleanUp(job *job.ProfilingJob) error {
	if recordingPIDs != nil && job.OutputType == api.Jfr {
		defer close(recordingPIDs)
		for _, pid := range j.targetPIDs {
			j.cleanUp(job, pid)
		}
	}

	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix+string(job.OutputType))

	return nil
}

func (j *jcmdManager) cleanUp(job *job.ProfilingJob, pid string) {
	recordingPIDs <- pid
	time.Sleep(1 * time.Second)
	cmd := jcmdStopCommand(j.commander, job, pid)
	if cmd != nil {
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.WarningLogLn(stderr.String())
		}
		_, _ = fmt.Fprint(io.Discard, out.String())
	}
}
