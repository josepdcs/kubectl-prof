package profiler

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/pkg/agent/config"

	"github.com/josepdcs/kubectl-prof/api"
)

type Profiler interface {
	SetUp(job *config.ProfilingJob) error
	Invoke(job *config.ProfilingJob) error
	CleanUp(job *config.ProfilingJob) error
}

func Get(lang api.ProgrammingLanguage, tool api.ProfilingTool) (Profiler, error) {
	switch lang {
	case api.Java:
		return NewJvmProfiler(), nil
	case api.Go, api.Node, api.Clang, api.ClangPlusPlus:
		if tool == api.Perf {
			return NewPerfProfiler(), nil
		}
		return NewBpfProfiler(), nil
	case api.Python:
		return NewPythonProfiler(), nil
	case api.Ruby:
		return NewRubyProfiler(), nil
	default:
		return nil, fmt.Errorf("could not find profiler for language %s", lang)
	}
}
