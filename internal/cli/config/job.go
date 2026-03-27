package config

import (
	"strings"

	apiv1 "k8s.io/api/core/v1"
)

// JobConfig holds configuration options for the profiling job that is launched
// by cli.
type JobConfig struct {

	// ContainerConfig holds the container spec
	ContainerConfig

	// Namespace specifies the namespace for job execution.
	Namespace string

	// Tolerations specifies the tolerations for the profiling job pod.
	Tolerations []apiv1.Toleration

	// TolerationsRaw holds raw toleration strings from command line
	TolerationsRaw []string
}

// DeepCopy returns a deep copy of the JobConfig.
func (j *JobConfig) DeepCopy() *JobConfig {
	tolerations := make([]apiv1.Toleration, len(j.Tolerations))
	copy(tolerations, j.Tolerations)
	tolerationsRaw := make([]string, len(j.TolerationsRaw))
	copy(tolerationsRaw, j.TolerationsRaw)
	return &JobConfig{
		ContainerConfig: j.ContainerConfig,
		Namespace:       j.Namespace,
		Tolerations:     tolerations,
		TolerationsRaw:  tolerationsRaw,
	}
}

// ParseTolerations parses raw toleration strings into apiv1.Toleration objects.
// Expected formats:
// - key=value:effect (e.g., "node.kubernetes.io/disk-pressure=true:NoSchedule")
// - key:effect (e.g., "node.kubernetes.io/memory-pressure:NoExecute")
// - key (e.g., "node.kubernetes.io/unreachable") - defaults to NoSchedule
func (j *JobConfig) ParseTolerations() error {
	if len(j.TolerationsRaw) == 0 {
		return nil
	}

	j.Tolerations = make([]apiv1.Toleration, 0, len(j.TolerationsRaw))

	for _, raw := range j.TolerationsRaw {
		toleration := apiv1.Toleration{
			Operator: apiv1.TolerationOpEqual,
		}

		// Split by colon to separate key[=value] from effect
		parts := strings.SplitN(raw, ":", 2)

		// Parse key and optionally value
		keyValue := parts[0]
		if strings.Contains(keyValue, "=") {
			kvParts := strings.SplitN(keyValue, "=", 2)
			toleration.Key = kvParts[0]
			toleration.Value = kvParts[1]
		} else {
			toleration.Key = keyValue
			toleration.Operator = apiv1.TolerationOpExists
		}

		// Parse effect if provided
		if len(parts) > 1 {
			toleration.Effect = apiv1.TaintEffect(parts[1])
		} else {
			toleration.Effect = apiv1.TaintEffectNoSchedule
		}

		j.Tolerations = append(j.Tolerations, toleration)
	}

	return nil
}
