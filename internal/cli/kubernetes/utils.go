package kubernetes

import (
	"strconv"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
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
		"--target-container-runtime", string(cfg.Target.ContainerRuntime),
		"--target-container-runtime-path", cfg.Target.ContainerRuntimePath,
		"--target-pod-uid", string(targetPod.UID),
		"--target-container-id", cfg.Target.ContainerID,
		"--lang", string(cfg.Target.Language),
		"--event-type", string(cfg.Target.Event),
		"--compressor-type", string(cfg.Target.Compressor),
		"--profiling-tool", string(cfg.Target.ProfilingTool),
		"--output-type", string(cfg.Target.OutputType),
		"--grace-period-ending", cfg.Target.GracePeriodEnding.String(),
	}

	args = appendArgument(args, "--job-id", id, func() bool { return stringUtils.IsNotBlank(id) })
	args = appendArgument(args, "--duration", cfg.Target.Duration.String(), func() bool { return cfg.Target.Duration > 0 })
	args = appendArgument(args, "--interval", cfg.Target.Interval.String(), func() bool { return cfg.Target.Interval > 0 })
	args = appendArgument(args, "--print-logs", "", func() bool { return cfg.Target.PrintLogs })
	args = appendArgument(args, "--heap-dump-split-in-chunk-size", cfg.Target.HeapDumpSplitInChunkSize, func() bool { return cfg.Target.OutputType == api.HeapDump || cfg.Target.OutputType == api.HeapSnapshot })
	args = appendArgument(args, "--pid", cfg.Target.PID, func() bool { return stringUtils.IsNotBlank(cfg.Target.PID) })
	args = appendArgument(args, "--pgrep", cfg.Target.Pgrep, func() bool { return stringUtils.IsNotBlank(cfg.Target.Pgrep) })
	args = appendArgument(args, "--node-heap-snapshot-signal", strconv.Itoa(cfg.Target.NodeHeapSnapshotSignal), func() bool {
		return cfg.Target.NodeHeapSnapshotSignal > 0 && cfg.Target.Language == api.Node
	})

	return args
}

func appendArgument(args []string, key string, value string, condition func() bool) []string {
	if condition() {
		if stringUtils.IsNotBlank(value) {
			args = append(args, key, value)
		} else {
			args = append(args, key)
		}
	}
	return args
}
