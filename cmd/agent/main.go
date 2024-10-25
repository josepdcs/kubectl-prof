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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/action/profile"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/urfave/cli/v2"
)

// gracePeriod is the default grace period so that the cli will be able to retrieve the result file,
// and also used in case of error for remaining before definitely ending the agent
var gracePeriod = 5 * time.Minute

// p profiler to be run
var p profiler.Profiler

// profilingJob the running profiling job
var profilingJob *job.ProfilingJob

// global err to be return
var err error

// done global channel which will be used as a flag for marking the profiling as finished
var done chan bool

func main() {
	done = make(chan bool, 1)

	// handle for TERM signal
	handleSignals()

	// handle for done channel in order to finish app
	handleForDone()

	// run main app, any error is logged
	err = runApp()
	if err != nil {
		log.ErrorLn(err)
	}

	// handle for app ending
	handleForEnding()
}

// runApp runs the agent
func runApp() error {
	app := &cli.App{
		Name:        "agent",
		UsageText:   "agent [global options]",
		Usage:       "the agent profiler used by kubectl-prof",
		Description: "An agent with capability for profiling containers inside pods",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     profile.JobId,
				Usage:    "job ID",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.TargetContainerRuntime,
				Usage:    "target container runtime (crio, containerd)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.TargetContainerRuntimePath,
				Usage:    "target container runtime path",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.TargetPodUID,
				Usage:    "target pod UID",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.TargetContainerID,
				Usage:    "target container ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     profile.Duration,
				Usage:    "profiling session duration",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.Interval,
				Usage:    "profiling interval",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.Lang,
				Usage:    "programming language",
				Required: true,
			},
			&cli.StringFlag{
				Name:     profile.EventType,
				Usage:    "profiling event type",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.CompressorType,
				Usage:    "compressor algorithm type",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.ProfilingTool,
				Usage:    "tool for profiling or debugging",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.OutputType,
				Usage:    "output type",
				Required: true,
			},
			&cli.StringFlag{
				Name:     profile.Filename,
				Usage:    "result file name",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     profile.PrintLogs,
				Usage:    "print logs",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.GracePeriodForEnding,
				Usage:    "grace period for agent ending",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.HeapDumpSplitInChunkSize,
				Usage:    "size of the chunks used to split the heap dump",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.Pid,
				Usage:    "the PID of the target process to be profiled",
				Required: false,
			},
			&cli.StringFlag{
				Name:     profile.Pgrep,
				Usage:    "the name of the process to be profiled",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			period, errParse := time.ParseDuration(c.String(profile.GracePeriodForEnding))
			if errParse == nil {
				gracePeriod = period
			}

			var err error
			p, profilingJob, err = profile.NewAction(toArgs(c))
			if err != nil {
				return err
			}

			return profile.Run(p, profilingJob)
		},
	}

	return app.Run(os.Args)
}

func toArgs(c *cli.Context) map[string]interface{} {

	return map[string]interface{}{
		profile.JobId:                      c.String(profile.JobId),
		profile.TargetContainerRuntime:     c.String(profile.TargetContainerRuntime),
		profile.TargetContainerRuntimePath: c.String(profile.TargetContainerRuntimePath),
		profile.TargetPodUID:               c.String(profile.TargetPodUID),
		profile.TargetContainerID:          c.String(profile.TargetContainerID),
		profile.Duration:                   c.String(profile.Duration),
		profile.Interval:                   c.String(profile.Interval),
		profile.Lang:                       c.String(profile.Lang),
		profile.EventType:                  c.String(profile.EventType),
		profile.CompressorType:             c.String(profile.CompressorType),
		profile.ProfilingTool:              c.String(profile.ProfilingTool),
		profile.OutputType:                 c.String(profile.OutputType),
		profile.Filename:                   c.String(profile.Filename),
		profile.PrintLogs:                  c.Bool(profile.PrintLogs),
		profile.GracePeriodForEnding:       c.String(profile.GracePeriodForEnding),
		profile.HeapDumpSplitInChunkSize:   c.String(profile.HeapDumpSplitInChunkSize),
		profile.Pid:                        c.String(profile.Pid),
		profile.Pgrep:                      c.String(profile.Pgrep),
	}
}

// handleSignals handles SIGTERM for ending the app
func handleSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		s := <-sigs
		log.DebugLogLn(fmt.Sprintf("Received signal: %s", s))
		if p != nil {
			err := p.CleanUp(profilingJob)
			if err != nil {
				log.ErrorLogLn(err.Error())
			}
		}

		done <- true
	}()
}

func handleForDone() {
	go func() {
		for {
			select {
			case <-done:
				log.DebugLogLn("Profiling finished properly. Bye!")
				os.Exit(0)
			default:
				// nothing to do
			}
		}
	}()
}

// handleForEnding handles the end of the app.
// A grace of period is defined so that the client profiler have time to retrieve the result of the profiling.
// Passed this time, the agent will be auto-deleted.
func handleForEnding() {
	timer := time.NewTimer(gracePeriod)
	fired := make(chan bool, 1)
	go func() {
		<-timer.C
		fired <- true
	}()

	for {
		select {
		case <-fired:
			log.WarningLogLn(fmt.Sprintf("Maximum allowed time %s surpassed. Cleaning up and auto-deleting the agent...", gracePeriod.String()))
			if p != nil {
				err := p.CleanUp(profilingJob)
				if err != nil {
					log.ErrorLogLn(err.Error())
				}
			}
			return
		default:
			// nothing to do
		}
	}
}
