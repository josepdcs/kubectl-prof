package job

import (
	"errors"
	"github.com/josepdcs/kubectl-prof/pkg/cli/config"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"

	"github.com/josepdcs/kubectl-prof/api"
)

const (
	baseImageName = "josepdcs/kubectl-prof"
	ContainerName = "kubectl-prof"
	LabelID       = "kubectl-prof/id"
)

type Creator interface {
	Create(targetPod *apiv1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error)
}

func Get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	switch lang {
	case api.Java:
		return &jvmCreator{}, nil
	case api.Go:
		return &bpfCreator{}, nil
	case api.Python:
		return &pythonCreator{}, nil
	case api.Ruby:
		return &rubyCreator{}, nil
	case api.Node:
		if tool == api.Perf {
			return &perfCreator{}, nil
		}
		return &bpfCreator{}, nil
	}

	// Should not happen
	return nil, errors.New("got language without job creator")
}
