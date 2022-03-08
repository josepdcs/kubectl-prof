//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.
package cli

import (
	"context"
	"errors"
	"fmt"
	data2 "github.com/josepdcs/kubectl-profiling/internal/cli/data"
	"github.com/josepdcs/kubectl-profiling/internal/cli/handler"
	kubernetes2 "github.com/josepdcs/kubectl-profiling/internal/cli/kubernetes"
	"log"

	v1 "k8s.io/api/core/v1"
)

func Flame(cfg *data2.FlameConfig) {
	ns, err := kubernetes2.Connect(cfg.ConfigFlags)
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
	pod, err := kubernetes2.GetPodDetails(cfg.TargetConfig.PodName, cfg.TargetConfig.Namespace, ctx)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	containerName, err := validatePod(pod, cfg.TargetConfig)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	containerId, err := kubernetes2.GetContainerId(containerName, pod)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	p.PrintSuccess()

	cfg.TargetConfig.ContainerName = containerName
	cfg.TargetConfig.ContainerId = containerId

	p.Print("Launching profiler ... ")
	profileId, job, err := kubernetes2.LaunchFlameJob(pod, cfg, ctx)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	if cfg.TargetConfig.DryRun {
		return
	}

	cfg.TargetConfig.Id = profileId
	profilerPod, err := kubernetes2.WaitForPodStart(cfg, ctx)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	p.PrintSuccess()
	apiHandler := &handler.ApiEventsHandler{
		Job:    job,
		Target: cfg.TargetConfig,
	}
	done, err := kubernetes2.GetLogsFromPod(profilerPod, apiHandler, ctx)
	if err != nil {
		p.PrintError()
		fmt.Println(err.Error())
	}

	<-done
}

func validatePod(pod *v1.Pod, targetDetails *data2.TargetDetails) (string, error) {
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

		return "", errors.New(fmt.Sprintf("Could not determine container. please specify one of %v", containerNames))
	}

	return pod.Spec.Containers[0].Name, nil
}
