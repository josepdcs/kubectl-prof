package config

// JobConfig holds configuration options for the profiling job that is launched
// by cli.
type JobConfig struct {

	// ContainerConfig holds the container spec
	ContainerConfig

	// Namespace specifies the namespace for job execution.
	Namespace string
}
