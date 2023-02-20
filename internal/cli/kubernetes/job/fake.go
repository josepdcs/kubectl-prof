package job

import (
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

type fakeCreator struct {
}

func (m *fakeCreator) Create(targetPod *v1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error) {
	return "ID", &batchv1.Job{}, nil
}
