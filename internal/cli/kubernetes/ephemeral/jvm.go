package ephemeral

import (
	"fmt"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/version"
	v1 "k8s.io/api/core/v1"
)

type jvmCreator struct{}

func (j *jvmCreator) Create(targetPod *v1.Pod, cfg *config.ProfilerConfig) *v1.EphemeralContainer {
	imageName := j.getImageName(cfg.Target)

	ephemeralContainer := &v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			ImagePullPolicy: cfg.Target.ImagePullPolicy,
			Name:            getContainerName(),
			Image:           imageName,
			Command:         []string{command},
			Args:            kubernetes.GetArgs(targetPod, cfg, ""),
			SecurityContext: &v1.SecurityContext{
				Privileged: &cfg.EphemeralContainer.Privileged,
			},
		},
		TargetContainerName: cfg.Target.ContainerName,
	}

	return ephemeralContainer
}

func (c *jvmCreator) getImageName(t *config.TargetConfig) string {
	if t.Image != "" {
		return t.Image
	}

	tag := fmt.Sprintf("%s-jvm", version.GetCurrent())
	if t.Alpine {
		tag = fmt.Sprintf("%s-alpine", tag)
	}

	return fmt.Sprintf("%s:%s", baseImageName, tag)
}
