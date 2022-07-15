/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package profiler

import (
	"bytes"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/utils"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"
)

const (
	profilerDir = "/tmp/async-profiler"
	profilerSh  = profilerDir + "/profiler.sh"

	jcmd        = "/opt/jdk-17/bin/jcmd"
	jcmdMaxSize = "maxsize=100M"
)

var stopJcmdRecording chan bool

var jvmResultFile = func(job *config.ProfilingJob) string {
	if stringUtils.IsBlank(job.FileName) {
		switch job.OutputType {
		case api.Jfr:
			return "/tmp/flight.jfr"
		case api.ThreadDump:
			return "/tmp/threaddump.txt"
		case api.HeapDump:
			return "/tmp/heapdump.hprof"
		case api.HeapHistogram:
			return "/tmp/heaphistogram.txt"
		case api.Flat:
			return "/tmp/flat.txt"
		case api.Traces:
			return "/tmp/traces.txt"
		case api.Collapsed:
			return "/tmp/collapsed.txt"
		case api.Tree:
			return "/tmp/tree.html"
		default:
			return "/tmp/flamegraph.html"
		}
	}
	return "/tmp/" + job.FileName
}

var jvmCommand = func(job *config.ProfilingJob, pid string, fileName string) *exec.Cmd {
	duration := strconv.Itoa(int(job.Duration.Seconds()))
	if job.ProfilingTool == api.Jcmd {
		switch job.OutputType {
		case api.Jfr:
			return utils.Command(jcmd, pid, "JFR.start", jcmdMaxSize, "duration="+duration+"s", "filename="+fileName, "name=pid_"+pid)
		case api.ThreadDump:
			return utils.Command(jcmd, pid, "Thread.print")
		case api.HeapDump:
			return utils.Command(jcmd, pid, "GC.heap_dump", fileName)
		case api.HeapHistogram:
			return utils.Command(jcmd, pid, "GC.class_histogram")
		}
	}
	// async-profiler
	event := string(job.Event)
	output := string(job.OutputType)
	return utils.Command(profilerSh, "-o", output, "-d", duration, "-f", fileName, "-e", event, "--fdtransfer", pid)
}

var jvmStopCommand = func(job *config.ProfilingJob, pid string) *exec.Cmd {
	if job.ProfilingTool == api.Jcmd {
		return utils.Command(jcmd, pid, "JFR.stop", "name=pid_"+pid)
	}
	return utils.Command(profilerSh, "stop", pid)
}

type JvmProfiler struct {
	targetPID string
	JvmUtil
}

type JvmUtil interface {
	copyProfilerToTempDirIfNeeded(tool api.ProfilingTool) error
	handleProfilingResult(job *config.ProfilingJob, fileName string, out bytes.Buffer, targetPID string) error
	handleJcmdRecording(targetPID string)
	publishResult(compressor api.Compressor, fileName string, outputType api.EventType) error
}

type jvmUtil struct {
}

func NewJvmProfiler() *JvmProfiler {
	return &JvmProfiler{JvmUtil: &jvmUtil{}}
}

func (j *JvmProfiler) SetUp(job *config.ProfilingJob) error {
	targetFs, err := utils.ContainerFileSystem(job.ContainerRuntime, job.ContainerID)
	if err != nil {
		return err
	}
	utils.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The target filesystem is: %s", targetFs))

	err = os.RemoveAll("/tmp")
	if err != nil {
		return err
	}

	err = os.Symlink(path.Join(targetFs, "tmp"), "/tmp")
	if err != nil {
		return err
	}

	return j.copyProfilerToTempDirIfNeeded(job.ProfilingTool)
}

func (j *JvmProfiler) Invoke(job *config.ProfilingJob) error {
	pid, err := utils.ContainerPID(job, false)
	if err != nil {
		return err
	}
	utils.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The PID to be profiled: %s", pid))
	j.targetPID = pid

	var out bytes.Buffer
	var stderr bytes.Buffer

	fileName := jvmResultFile(job)
	cmd := jvmCommand(job, pid, fileName)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		utils.PublishLogEvent(api.ErrorLevel, out.String())
		utils.PublishLogEvent(api.ErrorLevel, stderr.String())
		return fmt.Errorf("could not launch profiler: %w", err)
	}

	err = j.handleProfilingResult(job, fileName, out, j.targetPID)
	if err != nil {
		return err
	}

	return j.publishResult(job.Compressor, fileName, job.OutputType)
}

func (j *JvmProfiler) CleanUp(job *config.ProfilingJob) error {
	if stopJcmdRecording != nil {
		stopJcmdRecording <- true
	}
	time.Sleep(2 * time.Second)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := jvmStopCommand(job, j.targetPID)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		utils.PublishLogEvent(api.WarnLevel, stderr.String())
	}
	_, _ = fmt.Fprint(io.Discard, out.String())

	err = os.RemoveAll("/tmp/async-profiler")
	if err != nil {
		utils.PublishLogEvent(api.WarnLevel, fmt.Sprintf("directory could no be removed: %s", err))
	}

	fileName := jvmResultFile(job)
	err = os.Remove(fileName + api.GetExtensionFileByCompressor[job.Compressor])
	if err != nil {
		utils.PublishLogEvent(api.WarnLevel, fmt.Sprintf("file could no be removed: %s", err))
	}

	return os.Remove(fileName)
}

func (j *jvmUtil) copyProfilerToTempDirIfNeeded(tool api.ProfilingTool) error {
	if tool == api.Jcmd {
		return nil
	}
	cmd := utils.Command("cp", "-r", "/app/async-profiler", "/tmp")
	return cmd.Run()
}

func (j *jvmUtil) handleProfilingResult(job *config.ProfilingJob, fileName string, out bytes.Buffer, targetPID string) error {
	if job.ProfilingTool == api.Jcmd {
		switch job.OutputType {
		case api.Jfr:
			j.handleJcmdRecording(targetPID)
		case api.ThreadDump, api.HeapHistogram:
			err := ioutil.WriteFile(fileName, out.Bytes(), 0644)
			if err != nil {
				return fmt.Errorf("could not save dump to file: %w", err)
			}
		default:
			utils.PublishLogEvent(api.DebugLevel, out.String())
		}
	} else {
		utils.PublishLogEvent(api.DebugLevel, out.String())
	}
	return nil
}

func (j *jvmUtil) handleJcmdRecording(targetPID string) {
	stopJcmdRecording = make(chan bool, 1)
	done := make(chan bool, 1)

	go func() {
		utils.PublishLogEvent(api.DebugLevel, "Jcmd recording is running right now, be patient...")
		for {
			select {
			case <-stopJcmdRecording:
				utils.PublishLogEvent(api.WarnLevel, "Stopping is detected. Ignoring jcmd recording...")
				return
			default:
				var out bytes.Buffer
				var stderr bytes.Buffer
				cmd := utils.SilentCommand(jcmd, targetPID, "JFR.check", "name=pid_"+targetPID)
				cmd.Stdout = &out
				cmd.Stderr = &stderr
				err := cmd.Run()
				if err != nil {
					utils.PublishLogEvent(api.ErrorLevel, stderr.String())
					done <- true
					return
				}
				if !stringUtils.Contains(out.String(), "running") {
					done <- true
					return
				}
			}
		}
	}()

	<-done
}

func (j *jvmUtil) publishResult(c api.Compressor, fileName string, outputType api.EventType) error {
	return utils.Publish(c, fileName, outputType)
}
