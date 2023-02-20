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

package profiler

import (
	"context"
	"errors"
	"fmt"
	"github.com/josepdcs/kubectl-prof/internal/cli"
	"github.com/josepdcs/kubectl-prof/internal/cli/adapter"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/handler"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type Profiler struct {
	profilingAdapter adapter.ProfilingAdapter
}

// NewProfiler returns new Profiler
func NewProfiler(profilingAdapter adapter.ProfilingAdapter) Profiler {
	return Profiler{
		profilingAdapter: profilingAdapter,
	}
}

// Profile runs all the steps of a profiling execution
func (p Profiler) Profile(cfg *config.ProfilerConfig) {
	ctx := context.Background()

	printer := cli.NewPrinter(cfg.Target.DryRun)

	printer.Print("Verifying target pod ... ")
	pod, err := p.profilingAdapter.GetTargetPod(cfg.Target.PodName, cfg.Target.Namespace, ctx)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	containerName, err := validatePod(pod, cfg.Target)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	containerId, err := kubernetes.ToContainerId(containerName, pod)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	printer.PrintSuccess()

	cfg.Target.ContainerName = containerName
	cfg.Target.ContainerId = containerId

	printer.Print("Launching profiler ... ")
	profileId, job, err := p.profilingAdapter.CreateProfilingJob(pod, cfg, ctx)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	if cfg.Target.DryRun {
		return
	}

	cfg.Target.Id = profileId
	profilingPod, err := p.profilingAdapter.GetProfilingPod(cfg, ctx)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	printer.PrintSuccess()
	eventHandler := handler.NewEventHandler(job, cfg.Target, cfg.LogLevel)
	done, resultFile, err := p.profilingAdapter.GetProfilingPodLogs(profilingPod, eventHandler, ctx)
	if err != nil {
		printer.PrintError()
		fmt.Println(err.Error())
	}

	var end bool
	for {
		select {
		case f := <-resultFile:
			fileName, err := p.profilingAdapter.GetRemoteFile(profilingPod, f, cfg.Target.LocalPath, cfg.Target.Compressor)
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
	_ = p.profilingAdapter.DeleteProfilingJob(job, ctx)
}

func validatePod(pod *v1.Pod, cfg *config.TargetConfig) (string, error) {
	if pod == nil {
		return "", errors.New(fmt.Sprintf("Could not find pod %s in Namespace %s",
			cfg.PodName, cfg.Namespace))
	}

	if len(pod.Spec.Containers) != 1 {
		var containerNames []string
		for _, container := range pod.Spec.Containers {
			if container.Name == cfg.ContainerName {
				return container.Name, nil // Found given container
			}

			containerNames = append(containerNames, container.Name)
		}

		return "", errors.New(fmt.Sprintf("Could not determine container. please specify one of %v",
			containerNames))
	}

	return pod.Spec.Containers[0].Name, nil
}
