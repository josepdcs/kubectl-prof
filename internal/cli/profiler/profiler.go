package profiler

import (
	"context"
	"fmt"
	"time"

	"github.com/alitto/pond"
	"github.com/josepdcs/kubectl-prof/internal/cli"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/handler"
	"github.com/josepdcs/kubectl-prof/internal/cli/profiler/api"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

// Profiler is a profiler job representation which wraps the api.PodApi, api.ProfilingJobApi
// and api.ProfilingContainerApi
type Profiler struct {
	podApi                api.PodApi
	profilingJobApi       api.ProfilingJobApi
	profilingContainerApi api.ProfilingContainerApi
}

// NewJobProfiler returns a new Profiler
func NewJobProfiler(podApi api.PodApi, profilingJobApi api.ProfilingJobApi,
	profilingContainerApi api.ProfilingContainerApi) *Profiler {
	return &Profiler{
		podApi:                podApi,
		profilingJobApi:       profilingJobApi,
		profilingContainerApi: profilingContainerApi,
	}
}

// Profile runs all the steps of the profiling from the job creation up to get the profiling result
func (p *Profiler) Profile(cfg *config.ProfilerConfig) error {
	if cfg.Target.PodName != "" {
		ctx := context.Background()
		printer := cli.NewPrinter(cfg.Target.DryRun)

		pod, err := p.podApi.GetPod(ctx, cfg.Target.PodName, cfg.Target.Namespace)
		if err != nil {
			return err
		}

		return p.profileTarget(ctx, pod, printer, cfg)
	}

	if cfg.Target.LabelSelector != "" {
		pods, err := p.podApi.GetPodsByLabelSelector(context.Background(), cfg.Target.Namespace, cfg.Target.LabelSelector)
		if err != nil {
			return err
		}

		if len(pods) == 0 {
			return errors.New(fmt.Sprintf("No pods found in namespace %s with label selector %s", cfg.Target.Namespace, cfg.Target.LabelSelector))
		}

		poolSize := cfg.Target.PoolSizeLaunchProfilingJobs
		if poolSize == 0 {
			poolSize = len(pods)
		}
		pool := pond.New(poolSize, 0, pond.MinWorkers(poolSize))
		defer pool.StopAndWait()

		// create a task group associated to a context
		group, _ := pool.GroupContext(context.Background())

		for _, pod := range pods {
			printer := cli.NewPrinterWithTargetPod(cfg.Target.DryRun, pod.Name)
			if pod.Status.Phase != v1.PodRunning {
				printer.Print(fmt.Sprintf("⚠️ Pod %s will be ignored because is not running, it is %s\n", pod.Name, pod.Status.Phase))
				continue
			}
			profilerConfig := cfg.DeepCopy()
			group.Submit(func() error {
				return p.profileTarget(context.Background(), &pod, printer, profilerConfig)
			})

		}

		return group.Wait()
	}

	return errors.New("no target specified")
}

// profileTarget runs all the steps of the profiling from the job creation
// up to get the profiling result for a target pod
func (p *Profiler) profileTarget(ctx context.Context, targetPod *v1.Pod, printer cli.Printer, cfg *config.ProfilerConfig) error {
	err := validatePodAndRetrieveContainerInfo(targetPod, cfg)
	if err != nil {
		return err
	}
	printer.Print("Verified target pod ... ✔\n")

	profileId, job, err := p.profilingJobApi.CreateProfilingJob(targetPod, cfg, ctx)
	if err != nil {
		return err
	}
	printer.Print("Launched profiler ... 🚀\n")

	if cfg.Target.DryRun {
		return nil
	}

	cfg.Target.Id = profileId
	profilingPod, err := p.profilingJobApi.GetProfilingPod(cfg, ctx, 5*time.Minute)
	if err != nil {
		return err
	}

	eventHandler := handler.NewEventHandler(cfg.Target, printer)
	done, resultFile, err := p.profilingContainerApi.HandleProfilingContainerLogs(profilingPod,
		p.profilingJobApi.GetProfilingContainerName(), eventHandler, ctx)
	if err != nil {
		return err
	}

	profilingStart := time.Now()
	var end bool
	for {
		select {
		case f := <-resultFile:
			start := time.Now()
			fileName, err := p.profilingContainerApi.GetRemoteFile(profilingPod, p.profilingJobApi.GetProfilingContainerName(), f, targetPod.Name, cfg.Target)
			if err != nil {
				printer.PrintError()
				fmt.Println(err.Error())
			} else {
				// downloaded result file
				elapsed := time.Since(start)
				printer.Print(fmt.Sprintf("Remote profiling file downloaded in %f seconds. ✔\n", elapsed.Seconds()))

				elapsed = time.Since(profilingStart)
				printer.Print(fmt.Sprintf("The profiling result file [%s] was obtained in %f seconds. 🔥\n", fileName, elapsed.Seconds()))
			}
		case end = <-done:
		}
		if end {
			break
		}
	}

	// invoke delete profiling job
	return p.profilingJobApi.DeleteProfilingJob(job, ctx)
}
