package main

import (
	"errors"
	"github.com/josepdcs/kubectl-profiling/internal/agent/details"
	"github.com/josepdcs/kubectl-profiling/internal/agent/profiler"
	"github.com/josepdcs/kubectl-profiling/internal/agent/utils"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josepdcs/kubectl-profiling/api"
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

func profilingJob() (*details.ProfilingJob, error) {
	if len(os.Args) != 9 && len(os.Args) != 10 {
		return nil, errors.New("expected 8 or 9 arguments")
	}

	duration, err := time.ParseDuration(os.Args[5])
	if err != nil {
		return nil, err
	}

	job := &details.ProfilingJob{}
	job.ID = os.Args[1]
	job.PodUID = os.Args[2]
	job.ContainerName = os.Args[3]
	job.ContainerID = utils.NormalizeContainerID(os.Args[4])
	job.Duration = duration
	job.Language = api.ProgrammingLanguage(os.Args[6])
	job.Event = api.ProfilingEvent(os.Args[7])
	job.ContainerRuntime = api.ContainerRuntime(os.Args[8])
	if len(os.Args) == 10 {
		job.TargetProcessName = os.Args[9]
	}

	return job, nil
}

func handleSignals() chan bool {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		<-sigs
		os.RemoveAll("/tmp/async-profiler")
		os.Remove("/tmp")
		done <- true
	}()

	return done
}

func handleError(err error) {
	if err != nil {
		api.PublishError(err)
		os.Exit(1)
	}
}
