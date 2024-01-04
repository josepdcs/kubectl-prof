package profiler

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/jvm"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"time"
)

type Profiler interface {
	SetUp(job *job.ProfilingJob) error
	Invoke(job *job.ProfilingJob) (error, time.Duration)
	CleanUp(job *job.ProfilingJob) error
}

func Get(tool api.ProfilingTool) Profiler {
	switch tool {
	case api.Jcmd:
		return jvm.NewJcmdProfiler()
	case api.AsyncProfiler:
		return jvm.NewAsyncProfiler()
	case api.Bpf:
		return NewBpfProfiler(executil.NewCommander(), publish.NewPublisher())
	case api.Pyspy:
		return NewPythonProfiler(executil.NewCommander(), publish.NewPublisher())
	case api.Perf:
		return NewPerfProfiler()
	case api.Rbspy:
		return NewRubyProfiler(executil.NewCommander(), publish.NewPublisher())
	default:
		// util for tests
		return NewMockProfiler()
	}
}
