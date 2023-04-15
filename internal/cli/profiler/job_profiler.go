package profiler

import (
	"context"
	"fmt"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/cli"
	"github.com/josepdcs/kubectl-prof/internal/cli/adapter"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/handler"
)

// JobProfiler is a profiler job representation which wraps the adapter.PodAdapter, adapter.ProfilingJobAdapter
// and adapter.ProfilingContainerAdapter
type JobProfiler struct {
	podAdapter                adapter.PodAdapter
	profilingJobAdapter       adapter.ProfilingJobAdapter
	profilingContainerAdapter adapter.ProfilingContainerAdapter
}

// NewJobProfiler returns a new JobProfiler
func NewJobProfiler(podAdapter adapter.PodAdapter, profilingJobAdapter adapter.ProfilingJobAdapter,
	profilingContainerAdapter adapter.ProfilingContainerAdapter) *JobProfiler {
	return &JobProfiler{
		podAdapter:                podAdapter,
		profilingJobAdapter:       profilingJobAdapter,
		profilingContainerAdapter: profilingContainerAdapter,
	}
}

// Profile runs all the steps of the profiling from the job creation up to obtain the profiling result
func (p *JobProfiler) Profile(cfg *config.ProfilerConfig) error {
	ctx := context.Background()

	printer := cli.NewPrinter(cfg.Target.DryRun)

	printer.Print("Verifying target pod ... ")
	pod, err := p.podAdapter.GetPod(cfg.Target.PodName, cfg.Target.Namespace, ctx)
	if err != nil {
		return err
	}

	err = validatePodAndRetrieveContainerInfo(pod, cfg)
	if err != nil {
		return err
	}
	printer.PrintSuccess()

	printer.Print("Launching profiler ... ")
	profileId, job, err := p.profilingJobAdapter.CreateProfilingJob(pod, cfg, ctx)
	if err != nil {
		return err
	}

	if cfg.Target.DryRun {
		return nil
	}

	cfg.Target.Id = profileId
	profilingPod, err := p.profilingJobAdapter.GetProfilingPod(cfg, ctx, 5*time.Minute)
	if err != nil {
		return err
	}

	printer.PrintSuccess()
	eventHandler := handler.NewEventHandler(cfg.Target, cfg.LogLevel)
	done, resultFile, err := p.profilingContainerAdapter.HandleProfilingContainerLogs(profilingPod,
		p.profilingJobAdapter.GetProfilingContainerName(), eventHandler, ctx)
	if err != nil {
		return err
	}

	var end bool
	for {
		select {
		case f := <-resultFile:
			fileName, err := p.profilingContainerAdapter.GetRemoteFile(profilingPod,
				p.profilingJobAdapter.GetProfilingContainerName(), f, cfg.Target.LocalPath, cfg.Target.Compressor)
			if err != nil {
				printer.PrintError()
				fmt.Println(err.Error())
			} else {
				// retrieved result file
				fmt.Printf("✔\nResult profiling data saved to: %s 🔥\n", fileName)
			}
		case end = <-done:
		}
		if end {
			break
		}
	}

	// invoke delete profiling job
	return p.profilingJobAdapter.DeleteProfilingJob(job, ctx)
}
