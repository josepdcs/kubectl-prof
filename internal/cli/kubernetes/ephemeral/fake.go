package ephemeral

import (
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	v1 "k8s.io/api/core/v1"
)

// fakeCreator fake implementation of Creator
type fakeCreator struct {
}

func (m *fakeCreator) Create(targetPod *v1.Pod, cfg *config.ProfilerConfig) *v1.EphemeralContainer {
	return &v1.EphemeralContainer{
		EphemeralContainerCommon: v1.EphemeralContainerCommon{
			Name:  "EphemeralContainerName",
			Image: "EphemeralContainerImage",
		},
		TargetContainerName: "TargetContainerName",
	}
}
