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

package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/josepdcs/kubectl-profile/internal/cli/config"
	"github.com/josepdcs/kubectl-profile/internal/cli/handler"
	"github.com/josepdcs/kubectl-profile/internal/cli/kubernetes"
	"log"

	v1 "k8s.io/api/core/v1"
)

func Flame(cfg *config.ProfilerConfig) {
	ns, err := kubernetes.Connect(cfg.ConfigFlags)
	if err != nil {
		log.Fatalf("Failed connecting to kubernetes cluster: %v\n", err)
	}

	p := NewPrinter(cfg.TargetConfig.DryRun)

	if cfg.TargetConfig.Namespace == "" {
		cfg.TargetConfig.Namespace = ns
	}
	cfg.JobConfig.Namespace = ns

	ctx := context.Background()

	p.Print("Verifying target pod ... ")
	pod, err := kubernetes.GetPodDetails(cfg.TargetConfig.PodName, cfg.TargetConfig.Namespace, ctx)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	containerName, err := validatePod(pod, cfg.TargetConfig)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	containerId, err := kubernetes.GetContainerId(containerName, pod)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	p.PrintSuccess()

	cfg.TargetConfig.ContainerName = containerName
	cfg.TargetConfig.ContainerId = containerId

	p.Print("Launching profiler ... ")
	profileId, job, err := kubernetes.LaunchFlameJob(pod, cfg, ctx)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	if cfg.TargetConfig.DryRun {
		return
	}

	cfg.TargetConfig.Id = profileId
	profilerPod, err := kubernetes.WaitForPodStart(cfg, ctx)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	p.PrintSuccess()
	apiHandler := &handler.ApiEventsHandler{
		Job:    job,
		Target: cfg.TargetConfig,
	}
	done, err := kubernetes.GetLogsFromPod(profilerPod, apiHandler, ctx)
	if err != nil {
		p.PrintError()
		fmt.Println(err.Error())
	}

	<-done
}

func validatePod(pod *v1.Pod, targetDetails *config.TargetConfig) (string, error) {
	if pod == nil {
		return "", errors.New(fmt.Sprintf("Could not find pod %s in Namespace %s",
			targetDetails.PodName, targetDetails.Namespace))
	}

	if len(pod.Spec.Containers) != 1 {
		var containerNames []string
		for _, container := range pod.Spec.Containers {
			if container.Name == targetDetails.ContainerName {
				return container.Name, nil // Found given container
			}

			containerNames = append(containerNames, container.Name)
		}

		return "", errors.New(fmt.Sprintf("Could not determine container. please specify one of %v",
			containerNames))
	}

	return pod.Spec.Containers[0].Name, nil
}
