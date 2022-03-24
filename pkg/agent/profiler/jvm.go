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
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
	"os"
	"path"
	"strconv"
)

const (
	profilerDir = "/tmp/async-profiler"
	fileName    = "/tmp/flamegraph.html"
	profilerSh  = profilerDir + "/profiler.sh"
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

	return j.copyProfilerToTempDir()
}

func (j *JvmProfiler) Invoke(job *config.ProfilingJob) error {
	pid, err := utils.ContainerPID(job, false)
	if err != nil {
		return err
	}
	api.PublishLogEvent(api.DebugLevel, fmt.Sprintf("The PID to be profiled: %s", pid))

	duration := strconv.Itoa(int(job.Duration.Seconds()))
	event := string(job.Event)
	cmd := utils.Command(profilerSh, "-d", duration, "-f", fileName, "-e", event, pid)
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
	api.PublishLogEvent(api.InfoLevel, out.String())

	return utils.PublishFlameGraph(job.Compressor, fileName)
}

func (j *JvmProfiler) copyProfilerToTempDir() error {
	cmd := utils.Command("cp", "-r", "/app/async-profiler", "/tmp")
	return cmd.Run()
}
