package job

import (
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	apiv1 "k8s.io/api/core/v1"
)

func getArgs(targetPod *apiv1.Pod, cfg *config.ProfilerConfig, id string) []string {
	args := []string{
		"--job-id", id,
		"--target-container-runtime", string(cfg.Target.ContainerRuntime),
		"--target-pod-uid", string(targetPod.UID),
		"--target-container-id", cfg.Target.ContainerId,
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

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
