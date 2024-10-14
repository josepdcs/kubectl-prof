package ephemeral

import (
	"fmt"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/version"
	v1 "k8s.io/api/core/v1"
)

type bpfCreator struct{}

func (b *bpfCreator) Create(targetPod *v1.Pod, cfg *config.ProfilerConfig) *v1.EphemeralContainer {
	imageName := b.getImageName(cfg.Target)

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

// getImageName if image name is provider from config.TargetConfig this one is returned otherwise a new one is built
func (b *bpfCreator) getImageName(t *config.TargetConfig) string {
	var imageName string
	if t.Image != "" {
		imageName = t.Image
	} else {
		imageName = fmt.Sprintf("%s:%s-bpf", baseImageName, version.GetCurrent())
	}
	return imageName
}
