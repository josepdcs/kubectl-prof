package config

// EphemeralContainerConfig wraps the ephemeral container`s configuration to be launched.
type EphemeralContainerConfig struct {

	// Privileged indicates if ephemeral container has to be run in privileged mode.
	Privileged bool
}
