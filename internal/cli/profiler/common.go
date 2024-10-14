package profiler

import (
	"errors"
	"fmt"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	v1 "k8s.io/api/core/v1"
)

func validatePodAndRetrieveContainerInfo(pod *v1.Pod, cfg *config.ProfilerConfig) error {
	containerName, err := validatePod(pod, cfg.Target)
	if err != nil {
		return err
	}

	containerId, err := kubernetes.ToContainerId(containerName, pod)
	if err != nil {
		return err
	}

	cfg.Target.ContainerName = containerName
	cfg.Target.ContainerID = containerId

	return nil
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
