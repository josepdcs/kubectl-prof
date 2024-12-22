package config

import (
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerConfig wraps the container`s configuration to be launched.
type ContainerConfig struct {
	// RequestConfig configures resource requests for the job that is started.
	RequestConfig ResourceConfig

	// LimitConfig configures resource limits for the job that is started.
	LimitConfig ResourceConfig

	// Privileged indicates if Job has to be run in privileged mode
	Privileged bool

	// Capabilities indicate the capabilities that the container will have
	Capabilities []apiv1.Capability
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
		return out, errors.Wrap(err, "unable to generate container requests")
	}

	limits, err := c.LimitConfig.ParseResources()
	if err != nil {
		return out, errors.Wrap(err, "unable to generate container limits")
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
			return nil, errors.Wrapf(err, "unable to parse CPU value %q", rc.CPU)
		}

		list[apiv1.ResourceCPU] = cpu
	}

	if rc.Memory != "" {
		mem, err := resource.ParseQuantity(rc.Memory)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse memory value %q", rc.Memory)
		}

		list[apiv1.ResourceMemory] = mem
	}

	return list, nil
}
