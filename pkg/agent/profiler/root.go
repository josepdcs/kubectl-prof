package profiler

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"

	"github.com/josepdcs/kubectl-prof/api"
)

type Profiler interface {
	SetUp(job *config.ProfilingJob) error
	Invoke(job *config.ProfilingJob) error
}

var (
	jvm    = JvmProfiler{}
	bpf    = BpfProfiler{}
	python = PythonProfiler{}
	ruby   = RubyProfiler{}
	perf   = PerfProfiler{}
)

func Get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Profiler, error) {
	switch lang {
	case api.Java:
		return &jvm, nil
	case api.Go:
		return &bpf, nil
	case api.Python:
		return &python, nil
	case api.Ruby:
		return &ruby, nil
	case api.Node:
		if tool == api.Perf {
			return &perf, nil
		}
		return &bpf, nil
	default:
		return nil, fmt.Errorf("could not find profiler for language %s", lang)
	}
}
