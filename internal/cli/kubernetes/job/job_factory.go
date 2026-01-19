package job

import (
	"errors"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"

	"github.com/josepdcs/kubectl-prof/api"
)

const (
	command       = "/app/agent"
	baseImageName = "josepdcs/kubectl-prof"
	ContainerName = "kubectl-prof"
	LabelID       = "kubectl-prof/id"
)

// Creator defines the method for creating the profiling job according the programming language.
type Creator interface {
	Create(targetPod *apiv1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error)
}

// Get returns the Creator implementation according the programming language.
func Get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Creator, error) {
	switch lang {
	case api.Java:
		return &jvmCreator{}, nil
	case api.Go, api.Clang, api.ClangPlusPlus, api.Node:
		if tool == api.Perf {
			return &perfCreator{}, nil
		}
		if tool == api.NodeDummy {
			return &dummyCreator{}, nil
		}
		if tool == api.Btf {
			return &btfCreator{}, nil
		}
		return &bpfCreator{}, nil
	case api.Rust:
		if tool == api.CargoFlame {
			return &rustCreator{}, nil
		}
		if tool == api.Perf {
			return &perfCreator{}, nil
		}
		if tool == api.Btf {
			return &btfCreator{}, nil
		}
		return &bpfCreator{}, nil
	case api.Python:
		return &pythonCreator{}, nil
	case api.Ruby:
		return &rubyCreator{}, nil
	case api.FakeLang:
		return &fakeCreator{}, nil
	}

	// Should not happen
	return nil, errors.New("got language without job creator")
}
