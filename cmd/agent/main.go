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
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler"
	"github.com/josepdcs/kubectl-prof/internal/agent/utils"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
)

//gracePeriod grace period so that the cli will be able to retrieve the result file during this one
const gracePeriod = 5 * time.Minute

// p profiler to be run
var p profiler.Profiler

// profilingJob the running profiling job
var profilingJob = &config.ProfilingJob{}

// done global channel which will be used as a flag for marking the profiling as finished
var done chan bool

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
			duration, err := time.ParseDuration(c.String("duration"))
			if err != nil {
				return err
			}

			profilingJob.ID = c.String("job-id")
			profilingJob.PodUID = c.String("pod-uid")
			profilingJob.ContainerName = c.String("container-name")
			profilingJob.ContainerID = utils.NormalizeContainerID(c.String("container-id"))
			profilingJob.Duration = duration
			profilingJob.Language = api.ProgrammingLanguage(c.String("lang"))
			profilingJob.Event = api.ProfilingEvent(c.String("event-type"))
			profilingJob.ContainerRuntime = api.ContainerRuntime(c.String("container-runtime"))
			profilingJob.Compressor = api.Compressor(c.String("compressor"))
			profilingJob.ProfilingTool = api.ProfilingTool(c.String("tool"))
			profilingJob.OutputType = api.EventType(c.String("output-type"))
			profilingJob.FileName = c.String("filename")
			profilingJob.TargetProcessName = c.String("target-process")

			utils.PublishLogEvent(api.DebugLevel, profilingJob.String())

			err = utils.PublishEvent(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Started})
			if err != nil {
				return err
			}

			p, err = profiler.Get(profilingJob.Language, profilingJob.ProfilingTool)
			if err != nil {
				return err
			}

			err = p.SetUp(profilingJob)
			if err != nil {
				return err
			}

			handleSignals()
			handleForDone()
			err = p.Invoke(profilingJob)
			if err != nil {
				return err
			}

			err = utils.PublishEvent(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Ended})
			if err != nil {
				return err
			}

			return handleForHappyEnding()
		},
	}

	return app.Run(os.Args)
}

// handleForHappyEnding handles the ending up of the app.
// A grace of period is defined so that the client profiler have time to retrieve the result of the profiling.
// Passed this time, the agent will be auto-deleted.
func handleForHappyEnding() error {
	timer := time.NewTimer(gracePeriod)
	fired := make(chan bool, 1)
	go func() {
		<-timer.C
		fired <- true
	}()

	for {
		select {
		case <-fired:
			utils.PublishLogEvent(api.WarnLevel, "Maximum allowed time surpassed. Cleaning up and auto-deleting the agent...")
			return p.CleanUp(profilingJob)
		default:
			// nothing to do
		}
	}
}

// handleSignals handles SIGTERM for ending up the app
func handleSignals() {
	sigs := make(chan os.Signal, 1)
	done = make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		s := <-sigs
		utils.PublishLogEvent(api.DebugLevel, fmt.Sprintf("Recived signal: %s", s))
		err := p.CleanUp(profilingJob)
		if err != nil {
			utils.PublishLogEvent(api.WarnLevel, fmt.Sprintf("unexpected behaviour in cleanup: %s", err))
		}

		done <- true
	}()

}

func handleForDone() {
	go func() {
		for {
			select {
			case <-done:
				utils.PublishLogEvent(api.DebugLevel, "Bye!")
				os.Exit(0)
			default:
				// nothing to do
			}
		}
	}()
}

// handleError simple func helper for logging error and exit
func handleError(err error) {
	if err != nil {
		utils.PublishError(err)
		os.Exit(1)
	}
}
