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

	"github.com/josepdcs/kubectl-prof/internal/agent/action"
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
				Name:     action.JobId,
				Usage:    "job ID",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.TargetContainerRuntime,
				Usage:    "target container runtime (crio, containerd)",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.TargetContainerRuntimePath,
				Usage:    "target container runtime path",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.TargetPodUID,
				Usage:    "target pod UID",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.TargetContainerID,
				Usage:    "target container ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     action.Duration,
				Usage:    "profiling session duration",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.Interval,
				Usage:    "profiling interval",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.Lang,
				Usage:    "programming language",
				Required: true,
			},
			&cli.StringFlag{
				Name:     action.EventType,
				Usage:    "profiling event type",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.CompressorType,
				Usage:    "compressor algorithm type",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.ProfilingTool,
				Usage:    "tool for profiling or debugging",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.OutputType,
				Usage:    "output type",
				Required: true,
			},
			&cli.StringFlag{
				Name:     action.Filename,
				Usage:    "result file name",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     action.PrintLogs,
				Usage:    "print logs",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.GracePeriodForEnding,
				Usage:    "grace period for agent ending",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.OutputSplitInChunkSize,
				Usage:    "size of the chunks used to split the output file",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.Pid,
				Usage:    "the PID of the target process to be profiled",
				Required: false,
			},
			&cli.StringFlag{
				Name:     action.Pgrep,
				Usage:    "the name of the process to be profiled",
				Required: false,
			},
			&cli.IntFlag{
				Name:     action.NodeHeapSnapshotSignal,
				Usage:    "the signal to be sent to the target process to trigger a heap snapshot",
				Required: false,
			},
			&cli.StringSliceFlag{
				Name:     action.AsyncProfilerArg,
				Usage:    "additional arguments to pass to async-profiler",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			period, errParse := time.ParseDuration(c.String(action.GracePeriodForEnding))
			if errParse == nil {
				gracePeriod = period
			}

			var err error
			p, profilingJob, err = action.NewProfile(toArgs(c))
			if err != nil {
				return err
			}

			return action.Run(p, profilingJob)
		},
	}

	return app.Run(os.Args)
}

func toArgs(c *cli.Context) map[string]interface{} {

	return map[string]interface{}{
		action.JobId:                      c.String(action.JobId),
		action.TargetContainerRuntime:     c.String(action.TargetContainerRuntime),
		action.TargetContainerRuntimePath: c.String(action.TargetContainerRuntimePath),
		action.TargetPodUID:               c.String(action.TargetPodUID),
		action.TargetContainerID:          c.String(action.TargetContainerID),
		action.Duration:                   c.String(action.Duration),
		action.Interval:                   c.String(action.Interval),
		action.Lang:                       c.String(action.Lang),
		action.EventType:                  c.String(action.EventType),
		action.CompressorType:             c.String(action.CompressorType),
		action.ProfilingTool:              c.String(action.ProfilingTool),
		action.OutputType:                 c.String(action.OutputType),
		action.Filename:                   c.String(action.Filename),
		action.PrintLogs:                  c.Bool(action.PrintLogs),
		action.GracePeriodForEnding:       c.String(action.GracePeriodForEnding),
		action.OutputSplitInChunkSize:     c.String(action.OutputSplitInChunkSize),
		action.Pid:                        c.String(action.Pid),
		action.Pgrep:                      c.String(action.Pgrep),
		action.NodeHeapSnapshotSignal:     c.Int(action.NodeHeapSnapshotSignal),
		action.AsyncProfilerArg:           c.StringSlice(action.AsyncProfilerArg),
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
