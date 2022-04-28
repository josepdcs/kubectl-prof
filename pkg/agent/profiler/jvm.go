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
	"github.com/fsnotify/fsnotify"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path"
	"strconv"
)

const (
	profilerDir = "/tmp/async-profiler"
	profilerSh  = profilerDir + "/profiler.sh"

	jcmd        = "/opt/jdk-17/bin/jcmd"
	jcmdMaxSize = "maxsize=100M"
)

type JvmProfiler struct{}

func (j *JvmProfiler) SetUp(job *config.ProfilingJob) error {
	targetFs, err := utils.ContainerFileSystem(job.ContainerRuntime, job.ContainerID)
	if err != nil {
		return err
	}
	api.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The target filesystem is: %s", targetFs))

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
	api.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))

	var cmd *exec.Cmd
	var fileName string
	switch job.ProfilingTool {
	case api.Jcmd:
		fileName = "/tmp/flight.jfr"
		cmd = utils.Command(jcmd, pid, "JFR.start", jcmdMaxSize, "duration="+duration+"s", "filename="+fileName)
	default:
		event := string(job.Event)
		fileName = "/tmp/flamegraph.html"
		if job.OutputType == api.Jfr {
			fileName = "/tmp/flight.jfr"
		}
		output := string(job.OutputType)
		cmd = utils.Command(profilerSh, "-o", output, "-d", duration, "-f", fileName, "-e", event, "--fdtransfer", pid)
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		api.PublishLogEvent(api.ErrorLevel, out.String())
		api.PublishLogEvent(api.ErrorLevel, stderr.String())
		return fmt.Errorf("could not launch profiler: %w", err)
	}
	api.PublishLogEvent(api.DebugLevel, out.String())

	if job.ProfilingTool == api.Jcmd {
		j.handleJcmdRecording(fileName)
	}

	return utils.Publish(job.Compressor, fileName, job.OutputType)
}

func (j *JvmProfiler) copyProfilerToTempDirIfNeeded(tool api.ProfilingTool) error {
	if tool == api.Jcmd {
		return nil
	}
	cmd := utils.Command("cp", "-r", "/app/async-profiler", "/tmp")
	return cmd.Run()
}

func (j *JvmProfiler) handleJcmdRecording(fileName string) {
	done := make(chan bool)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		api.PublishError(err)
		log.Error(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			return
		}
	}(watcher)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					api.PublishLogEvent(api.DebugLevel, fmt.Sprintf("modified file: %s", event.Name))
					f, err := os.Stat(event.Name)
					if err != nil {
						api.PublishError(err)
						api.PublishLogEvent(api.ErrorLevel, err.Error())
						return
					}
					if f.Size() > 0 {
						done <- true
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				api.PublishError(err)
				api.PublishLogEvent(api.ErrorLevel, err.Error())
			}
		}
	}()

	//err = watcher.Add(j.targetTmpDir + "/flight.jfr")
	err = watcher.Add(fileName)
	api.PublishLogEvent(api.DebugLevel, fmt.Sprintf("add watcher to file: %s", fileName))

	if err != nil {
		api.PublishError(err)
		api.PublishLogEvent(api.ErrorLevel, err.Error())
	}

	<-done
}
