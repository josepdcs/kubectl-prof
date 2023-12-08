package profiler

import (
	"context"
	"fmt"
	"github.com/josepdcs/kubectl-prof/internal/cli"
	"github.com/josepdcs/kubectl-prof/internal/cli/adapter"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/handler"
	"time"
)

// EphemeralProfiler is a profiler representation by using an ephemeral container
// which wraps the adapter.PodAdapter, adapter.ProfilingEphemeralContainerAdapter and adapter.ProfilingContainerAdapter
type EphemeralProfiler struct {
	podAdapter                         adapter.PodAdapter
	profilingEphemeralContainerAdapter adapter.ProfilingEphemeralContainerAdapter
	profilingContainerAdapter          adapter.ProfilingContainerAdapter
}

// NewEphemeralProfiler returns a new EphemeralProfiler
func NewEphemeralProfiler(podAdapter adapter.PodAdapter, profilingEphemeralContainerAdapter adapter.ProfilingEphemeralContainerAdapter,
	profilingContainerAdapter adapter.ProfilingContainerAdapter) EphemeralProfiler {
	return EphemeralProfiler{
		podAdapter:                         podAdapter,
		profilingEphemeralContainerAdapter: profilingEphemeralContainerAdapter,
		profilingContainerAdapter:          profilingContainerAdapter,
	}
}

// Profile runs all the steps of the profiling: from creating and adding the new ephemeral container to the target Pod
// up to obtain the profiling result
func (e EphemeralProfiler) Profile(cfg *config.ProfilerConfig) error {
	ctx := context.Background()

	printer := cli.NewPrinter(cfg.Target.DryRun)

	printer.Print("Verifying target pod ... ")
	pod, err := e.podAdapter.GetPod(cfg.Target.PodName, cfg.Target.Namespace, ctx)
	if err != nil {
		return err
	}

	err = validatePodAndRetrieveContainerInfo(pod, cfg)
	if err != nil {
		return err
	}
	printer.PrintSuccess()

	printer.Print("Launching profiler ... ")
	podWithEphemeral, err := e.profilingEphemeralContainerAdapter.AddEphemeralContainer(pod, cfg, ctx, 5*time.Minute)
	if err != nil {
		return err
	}

	if cfg.Target.DryRun {
		return nil
	}

	printer.PrintSuccess()
	eventHandler := handler.NewEventHandler(cfg.Target, cfg.LogLevel)
	done, resultFile, err := e.profilingContainerAdapter.HandleProfilingContainerLogs(podWithEphemeral,
		e.profilingEphemeralContainerAdapter.GetEphemeralContainerName(), eventHandler, ctx)
	if err != nil {
		return err
	}

	var end bool
	for {
		select {
		case f := <-resultFile:
			fileName, err := e.profilingContainerAdapter.GetRemoteFile(podWithEphemeral, e.profilingEphemeralContainerAdapter.GetEphemeralContainerName(), f, cfg.Target)
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

	return nil
}
