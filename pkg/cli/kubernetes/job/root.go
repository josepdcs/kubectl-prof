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

var (
	jvm    = jvmCreator{}
	bpf    = bpfCreator{}
	python = pythonCreator{}
	ruby   = rubyCreator{}
	perf   = perfCreator{}
)

type creator interface {
	create(targetPod *apiv1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error)
}

func Create(targetPod *apiv1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error) {
	switch cfg.Target.Language {
	case api.Java:
		return jvm.create(targetPod, cfg)
	case api.Go:
		return bpf.create(targetPod, cfg)
	case api.Python:
		return python.create(targetPod, cfg)
	case api.Ruby:
		return ruby.create(targetPod, cfg)
	case api.Node:
		return bpf.create(targetPod, cfg)
	case api.NodeWithPerf:
		return perf.create(targetPod, cfg)
	}

	// Should not happen
	return "", nil, errors.New("got language without job creator")
}
