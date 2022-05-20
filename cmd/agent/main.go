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
	"fmt"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"
	"github.com/josepdcs/kubectl-prof/pkg/agent/profiler"
	"github.com/josepdcs/kubectl-prof/pkg/agent/utils"
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
		Name:        "agent",
		UsageText:   "agent [global options]",
		Usage:       "the agent profiler used by kubectl-prof",
		Description: "An agent with capability for profiling containers inside pods",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "job-id",
				Usage:    "job ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "pod-uid",
				Usage:    "pod UID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "container-name",
				Usage:    "container name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "container-id",
				Usage:    "container ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "duration",
				Usage:    "profiling duration",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "lang",
				Usage:    "programming language",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "event-type",
				Usage:    "profiling event type",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "container-runtime",
				Usage:    "container runtime",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "compressor",
				Usage:    "compressor algorithm type",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "tool",
				Usage:    "tool for profiling or debugging",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output-type",
				Usage:    "output type",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "filename",
				Usage:    "result file name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "target-process",
				Usage:    "target process name",
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

			utils.PublishLogEvent(api.DebugLevel, job.String())

			err = utils.PublishEvent(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Started})
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

			done := handleSignals(p, job)
			err = p.Invoke(job)
			if err != nil {
				return err
			}

			err = utils.PublishEvent(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Ended})
			if err != nil {
				return err
			}

			<-done

			return nil
		},
	}

	return app.Run(os.Args)
}

func handleSignals(p profiler.Profiler, job *config.ProfilingJob) chan bool {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		s := <-sigs
		utils.PublishLogEvent(api.DebugLevel, fmt.Sprintf("Recived signal: %s", s))
		err := p.CleanUp(job)
		if err != nil {
			return
		}

		done <- true
	}()

	return done
}

func handleError(err error) {
	if err != nil {
		utils.PublishError(err)
		os.Exit(1)
	}
}
