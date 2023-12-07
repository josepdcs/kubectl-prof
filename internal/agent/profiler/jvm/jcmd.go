package jvm

import (
	"bytes"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
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
)

var stopJcmdRecording chan bool

var jcmdCommand = func(job *job.ProfilingJob, pid string, fileName string) *exec.Cmd {
	switch job.OutputType {
	case api.ThreadDump:
		args := []string{pid, "Thread.print"}
		return util.Command(jcmd, args...)
	case api.HeapDump:
		args := []string{pid, "GC.heap_dump", fileName}
		return util.Command(jcmd, args...)
	case api.HeapHistogram:
		args := []string{pid, "GC.class_histogram"}
		return util.Command(jcmd, args...)
	}

	// default api.Jfr
	interval := strconv.Itoa(int(job.Interval.Seconds()))
	args := []string{pid, "JFR.start", "duration=" + interval + "s", "filename=" + fileName, invocationName + pid + "_" + string(job.OutputType), "settings=" + jfrSettingsTmpFilePath}
	return util.Command(jcmd, args...)
}

var jcmdStopCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
	if job.OutputType != api.Jfr {
		return nil
	}
	return util.Command(jcmd, pid, "JFR.stop", invocationName+pid+"_"+string(job.OutputType))
}

type JcmdProfiler struct {
	targetPID string
	JcmdManager
}

type JcmdManager interface {
	removeTmpDir() error
	linkTmpDirToTargetTmpDir(string) error
	copyJfrSettingsToTmpDir() error
	handleProfilingResult(job *job.ProfilingJob, fileName string, out bytes.Buffer, targetPID string) error
	handleJcmdRecording(targetPID string, outputType string)
	publishResult(compressor compressor.Type, fileName string, outputType api.OutputType, heapDumpSplitInChunkSize string) error
}

type jcmdManager struct {
}

func NewJcmdProfiler() *JcmdProfiler {
	return &JcmdProfiler{JcmdManager: &jcmdManager{}}
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

	pid, err := util.ContainerPID(job)
	if err != nil {
		return err
	}
	log.DebugLogLn(fmt.Sprintf("The PID to be profiled: %s", pid))
	j.targetPID = pid

	return j.copyJfrSettingsToTmpDir()
}

func (j *JcmdProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()
	fileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := jcmdCommand(job, j.targetPID, fileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.ErrorLogLn(out.String())
		return fmt.Errorf("could not launch profiler: %w; detail: %s", err, stderr.String()), time.Since(start)
	}

	err = j.handleProfilingResult(job, fileName, out, j.targetPID)
	if err != nil {
		return err, time.Since(start)
	}

	return j.publishResult(job.Compressor, fileName, job.OutputType, job.HeapDumpSplitInChunkSize), time.Since(start)
}

func (j *JcmdProfiler) CleanUp(job *job.ProfilingJob) error {
	if stopJcmdRecording != nil {
		stopJcmdRecording <- true

		time.Sleep(1 * time.Second)
		cmd := jcmdStopCommand(job, j.targetPID)
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

	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix+string(job.OutputType))

	return nil
}

func (j *jcmdManager) removeTmpDir() error {
	return os.RemoveAll(common.TmpDir())
}

func (j *jcmdManager) linkTmpDirToTargetTmpDir(targetTmpDir string) error {
	return os.Symlink(targetTmpDir, common.TmpDir())
}

func (j *jcmdManager) copyJfrSettingsToTmpDir() error {
	cmd := util.Command("cp", jfrSettingsImageFilePath, common.TmpDir())
	return cmd.Run()
}

func (j *jcmdManager) handleProfilingResult(job *job.ProfilingJob, fileName string, out bytes.Buffer, targetPID string) error {
	switch job.OutputType {
	case api.Jfr:
		j.handleJcmdRecording(targetPID, string(job.OutputType))
	case api.ThreadDump, api.HeapHistogram:
		err := os.WriteFile(fileName, out.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("could not save dump to file: %w", err)
		}
	default:
		log.DebugLogLn(out.String())
	}

	return nil
}

func (j *jcmdManager) handleJcmdRecording(targetPID string, outputType string) {
	stopJcmdRecording = make(chan bool, 1)
	done := make(chan bool)

	go func() {
		log.DebugLogLn("Jcmd recording is running right now, be patient...")
		for {
			select {
			case <-stopJcmdRecording:
				log.WarningLogLn("Stopping is detected. Ignoring jcmd recording...")
				return
			default:
				var out bytes.Buffer
				var stderr bytes.Buffer
				cmd := util.SilentCommand(jcmd, targetPID, "JFR.check", "name=pid_"+targetPID+"_"+outputType)
				cmd.Stdout = &out
				cmd.Stderr = &stderr
				err := cmd.Run()
				if err != nil {
					log.ErrorLogLn(stderr.String())
					done <- true
					return
				}
				outputTxt := out.String()
				if !stringUtils.Contains(outputTxt, "running") &&
					!stringUtils.Contains(outputTxt, "delayed") {
					// fix: in case of any error in a previous profiling which made that jcmd remains stopped
					if stringUtils.Contains(outputTxt, "stopped") {
						log.ErrorLogLn("Jcmd remains stopped for unknown reason, please restart the pod to be profiled (the pod application)")
					}
					done <- true
					return
				}

			}
		}
	}()

	<-done
}

func (j *jcmdManager) publishResult(c compressor.Type, fileName string, outputType api.OutputType, heapDumpSplitInChunkSize string) error {
	if outputType == api.HeapDump {
		return util.PublishWithNativeGzipAndSplit(fileName, heapDumpSplitInChunkSize, outputType)
	}
	return util.Publish(c, fileName, outputType)
}
