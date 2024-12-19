package profiler

import (
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/jvm"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
)

// Profiler is the interface that wraps the basic profiling operations.
type Profiler interface {
	// SetUp prepare the environment for the profiling
	SetUp(job *job.ProfilingJob) error
	// Invoke starts the profiling
	Invoke(job *job.ProfilingJob) (error, time.Duration)
	// CleanUp cleans the environment after the profiling
	CleanUp(job *job.ProfilingJob) error
}

// Get returns the profiler for the given tool
func Get(tool api.ProfilingTool) Profiler {
	switch tool {
	case api.Jcmd:
		return jvm.NewJcmdProfiler(executil.NewCommander(), publish.NewPublisher())
	case api.AsyncProfiler:
		return jvm.NewAsyncProfiler(executil.NewCommander(), publish.NewPublisher())
	case api.Bpf:
		return NewBpfProfiler(executil.NewCommander(), publish.NewPublisher())
	case api.Pyspy:
		return NewPythonProfiler(executil.NewCommander(), publish.NewPublisher())
	case api.Perf:
		return NewPerfProfiler(executil.NewCommander(), publish.NewPublisher())
	case api.Rbspy:
		return NewRubyProfiler(executil.NewCommander(), publish.NewPublisher())
	case api.NodeDummy:
		return NewNodeDummyProfiler(publish.NewPublisher())
	default:
		// util for tests
		return NewMockProfiler()
	}
}
