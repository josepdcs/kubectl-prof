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

package main

import (
	"errors"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/profiler"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
)

func main() {
	job, err := profilingJob()
	handleError(err)

	err = api.PublishEvent(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Started})
	handleError(err)

	p, err := profiler.ForLanguage(job.Language)
	handleError(err)

	err = p.SetUp(job)
	handleError(err)

	done := handleSignals()
	err = p.Invoke(job)
	handleError(err)

	err = api.PublishEvent(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Ended})
	handleError(err)

	<-done
}

func profilingJob() (*config.ProfilingJob, error) {
	if len(os.Args) != 12 && len(os.Args) != 13 {
		return nil, errors.New("expected 11 or 12 arguments")
	}

	duration, err := time.ParseDuration(os.Args[5])
	if err != nil {
		return nil, err
	}

	job := &config.ProfilingJob{}
	job.ID = os.Args[1]
	job.PodUID = os.Args[2]
	job.ContainerName = os.Args[3]
	job.ContainerID = utils.NormalizeContainerID(os.Args[4])
	job.Duration = duration
	job.Language = api.ProgrammingLanguage(os.Args[6])
	job.Event = api.ProfilingEvent(os.Args[7])
	job.ContainerRuntime = api.ContainerRuntime(os.Args[8])
	job.Compressor = api.Compressor(os.Args[9])
	job.ProfilingTool = api.ProfilingTool(os.Args[10])
	job.OutputType = api.EventType(os.Args[11])
	if len(os.Args) == 13 {
		job.TargetProcessName = os.Args[12]
	}

	api.PublishLogEvent(api.DebugLevel, job.String())

	return job, nil
}

func handleSignals() chan bool {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		s := <-sigs
		log.Infof("Recived signal: %s", s)
		err := os.RemoveAll("/tmp/async-profiler")
		if err != nil {
			log.Warnf("directory could no be removed: %s", err)
		}
		err = os.Remove("/tmp")
		if err != nil {
			// log.Warnf("directory could no be removed: %s", err)
		}
		done <- true
	}()

	return done
}

func handleError(err error, step ...string) {
	if err != nil {
		log.Errorf("%s", step)
		api.PublishError(err)
		os.Exit(1)
	}
}
