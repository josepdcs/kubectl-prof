package kubernetes

import (
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
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

func GetArgs(targetPod *apiv1.Pod, cfg *config.ProfilerConfig, id string) []string {
	args := []string{
		"--job-id", id,
		"--target-container-runtime", string(cfg.Target.ContainerRuntime),
		"--target-pod-uid", string(targetPod.UID),
		"--target-container-id", cfg.Target.ContainerID,
		"--lang", string(cfg.Target.Language),
		"--event-type", string(cfg.Target.Event),
		"--compressor-type", string(cfg.Target.Compressor),
		"--profiling-tool", string(cfg.Target.ProfilingTool),
		"--output-type", string(cfg.Target.OutputType),
		"--grace-period-ending", cfg.Target.GracePeriodEnding.String(),
	}

	if cfg.Target.Duration > 0 {
		args = append(args, "--duration", cfg.Target.Duration.String())
	}
	if cfg.Target.Interval > 0 {
		args = append(args, "--interval", cfg.Target.Interval.String())
	}
	if cfg.Target.PrintLogs {
		args = append(args, "--print-logs")
	}

	return args
}
