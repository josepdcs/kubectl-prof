package job

import (
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/pkg/cli/config"
	apiv1 "k8s.io/api/core/v1"
	"path/filepath"
)

func getArgs(targetPod *apiv1.Pod, cfg *config.ProfilerConfig, id string) []string {
	args := []string{
		"--job-id", id,
		"--pod-uid", string(targetPod.UID),
		"--container-name", cfg.Target.ContainerName,
		"--container-id", cfg.Target.ContainerId,
		"--duration", cfg.Target.Duration.String(),
		"--lang", string(cfg.Target.Language),
		"--event-type", string(cfg.Target.Event),
		"--container-runtime", string(cfg.Target.ContainerRuntime),
		"--compressor", string(cfg.Target.Compressor),
		"--tool", string(cfg.Target.ProfilingTool),
		"--output-type", string(cfg.Target.OutputType),
		"--filename", stringUtils.SubstringAfterLast(cfg.Target.FileName, string(filepath.Separator)),
	}

	if cfg.Target.Pgrep != "" {
		args = append(args, "--target-process", cfg.Target.Pgrep)
	}

	return args
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
