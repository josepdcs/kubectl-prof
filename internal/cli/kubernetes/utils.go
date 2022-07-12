package kubernetes

import (
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
)

func ToContainerId(containerName string, pod *apiv1.Pod) (string, error) {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			return containerStatus.ContainerID, nil
		}
	}

	return "", errors.New("Could not find container id for " + containerName)
}
