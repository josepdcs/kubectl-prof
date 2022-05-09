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
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/profiler"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
)

func main() {
	handleError(runApp())
}

func runApp() error {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "job-id",
				Usage:    "Job ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "pod-uid",
				Usage:    "Pod UID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "container-name",
				Usage:    "Container name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "container-id",
				Usage:    "Container ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "duration",
				Usage:    "Profiling duration",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "lang",
				Usage:    "Programming language",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "event-type",
				Usage:    "Profiling event type",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "container-runtime",
				Usage:    "Container runtime",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "compressor",
				Usage:    "Compressor algorithm type",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "tool",
				Usage:    "Tool for profiling or debugging",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output-type",
				Usage:    "Output type",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "filename",
				Usage:    "Result file name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "target-process",
				Usage:    "Target process name",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			job := &config.ProfilingJob{}
			duration, err := time.ParseDuration(c.String("duration"))
			if err != nil {
				return err
			}

			job.ID = c.String("job-id")
			job.PodUID = c.String("pod-uid")
			job.ContainerName = c.String("container-name")
			job.ContainerID = utils.NormalizeContainerID(c.String("container-id"))
			job.Duration = duration
			job.Language = api.ProgrammingLanguage(c.String("lang"))
			job.Event = api.ProfilingEvent(c.String("event-type"))
			job.ContainerRuntime = api.ContainerRuntime(c.String("container-runtime"))
			job.Compressor = api.Compressor(c.String("compressor"))
			job.ProfilingTool = api.ProfilingTool(c.String("tool"))
			job.OutputType = api.EventType(c.String("output-type"))
			job.FileName = c.String("filename")
			job.TargetProcessName = c.String("target-process")

			api.PublishLogEvent(api.DebugLevel, job.String())

			err = api.PublishEvent(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Started})
			if err != nil {
				return err
			}

			p, err := profiler.Get(job.Language, job.ProfilingTool)
			if err != nil {
				return err
			}

			err = p.SetUp(job)
			if err != nil {
				return err
			}

			done := handleSignals()
			err = p.Invoke(job)
			if err != nil {
				return err
			}

			err = api.PublishEvent(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Ended})
			if err != nil {
				return err
			}

			<-done

			return nil
		},
	}

	return app.Run(os.Args)
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
