package config

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ContainerConfig struct {
	// RequestConfig configures resource requests for the job that is started.
	RequestConfig ResourceConfig

	// LimitConfig configures resource limits for the job that is started.
	LimitConfig ResourceConfig

	// Privileged indicates if Job has to be run in privileged mode
	Privileged bool
}

// ResourceConfig holds resource configuration for either requests or limits.
type ResourceConfig struct {
	CPU    string
	Memory string
}

// ToResourceRequirements parses ContainerConfig into an apiv1.ResourceRequirements
// map which can be passed to a container spec.
func (c *ContainerConfig) ToResourceRequirements() (apiv1.ResourceRequirements, error) {
	var out apiv1.ResourceRequirements

	requests, err := c.RequestConfig.ParseResources()
	if err != nil {
		return out, fmt.Errorf("unable to generate container requests: %w", err)
	}

	limits, err := c.LimitConfig.ParseResources()
	if err != nil {
		return out, fmt.Errorf("unable to generate container limits: %w", err)
	}

	out.Requests = requests
	out.Limits = limits

	return out, nil
}

// ParseResources parses the ResourceConfig and returns an apiv1.ResourceList
// which can be used in a apiv1.ResourceRequirements map.
func (rc ResourceConfig) ParseResources() (apiv1.ResourceList, error) {
	if rc.CPU == "" && rc.Memory == "" {
		return nil, nil
	}

	list := make(apiv1.ResourceList)

	if rc.CPU != "" {
		cpu, err := resource.ParseQuantity(rc.CPU)
		if err != nil {
			return nil, fmt.Errorf("unable to parse CPU value %q: %w", rc.CPU, err)
		}

		list[apiv1.ResourceCPU] = cpu
	}

	if rc.Memory != "" {
		mem, err := resource.ParseQuantity(rc.Memory)
		if err != nil {
			return nil, fmt.Errorf("unable to parse memory value %q: %w", rc.Memory, err)
		}

		list[apiv1.ResourceMemory] = mem
	}

	return list, nil
}
