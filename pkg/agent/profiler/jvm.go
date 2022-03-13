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
	"github.com/josepdcs/kubectl-profile/api"
	"github.com/josepdcs/kubectl-profile/pkg/agent/details"
	"github.com/josepdcs/kubectl-profile/pkg/agent/utils"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"
)

const (
	profilerDir = "/tmp/async-profiler"
	fileName    = "/tmp/flamegraph.html"
	profilerSh  = profilerDir + "/profiler.sh"
)

type JvmProfiler struct{}

func (j *JvmProfiler) SetUp(job *details.ProfilingJob) error {
	targetFs, err := utils.TargetFileSystemLocation(job.ContainerRuntime, job.ContainerID)
	if err != nil {
		return err
	}
	_ = api.PublishEvent(
		api.Log,
		&api.LogData{
			Time:  time.Now(),
			Level: api.InfoLevel,
			Msg:   fmt.Sprintf("The target filesystem is: %s", targetFs)},
	)

	err = os.RemoveAll("/tmp")
	if err != nil {
		return err
	}

	err = os.Symlink(path.Join(targetFs, "tmp"), "/tmp")
	if err != nil {
		return err
	}

	return j.copyProfilerToTempDir()
}

func (j *JvmProfiler) Invoke(job *details.ProfilingJob) error {
	pid, err := utils.FindProcessId(job)
	//log.Infof("The PID to be profiled: %s", pid)
	if err != nil {
		return err
	}

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	event := string(job.Event)
	cmd := exec.Command(profilerSh, "-d", duration, "-f", fileName, "-e", event, pid)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		outStr := out.String()
		if len(outStr) > 0 {
			_ = api.PublishEvent(
				api.Log,
				&api.LogData{
					Time:  time.Now(),
					Level: api.ErrorLevel,
					Msg:   fmt.Sprint(outStr)},
			)
		}
		errStr := stderr.String()
		if len(errStr) > 0 {
			_ = api.PublishEvent(
				api.Log,
				&api.LogData{
					Time:  time.Now(),
					Level: api.ErrorLevel,
					Msg:   fmt.Sprint(errStr)},
			)
		}
		return err
	}

	/*outStr := out.String()
	if outStr != "" {
		_ = api.PublishEvent(
			api.Log,
			&api.LogData{
				Time:  time.Now(),
				Level: api.InfoLevel,
				Msg:   fmt.Sprint(outStr)},
		)
	}*/

	return utils.PublishFlameGraph(fileName)
}

func (j *JvmProfiler) copyProfilerToTempDir() error {
	cmd := exec.Command("cp", "-r", "/app/async-profiler", "/tmp")
	return cmd.Run()
}
