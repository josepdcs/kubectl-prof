package profiler

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/jvm"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGet(t *testing.T) {
	type args struct {
		tool api.ProfilingTool
	}
	tests := []struct {
		name string
		tool api.ProfilingTool
		want Profiler
	}{
		{
			name: "should return jcmd profiler",
			tool: api.Jcmd,
			want: jvm.NewJcmdProfiler(),
		},
		{
			name: "should return async profiler profiler",
			tool: api.AsyncProfiler,
			want: jvm.NewAsyncProfiler(),
		},
		{
			name: "should return bpf profiler",
			tool: api.Bpf,
			want: NewBpfProfiler(executil.NewCommander(), publish.NewPublisher()),
		},
		{
			name: "should return pyspy profiler",
			tool: api.Pyspy,
			want: NewPythonProfiler(executil.NewCommander(), publish.NewPublisher()),
		},
		{
			name: "should return Rbspy profiler",
			tool: api.Rbspy,
			want: NewRubyProfiler(executil.NewCommander(), publish.NewPublisher()),
		},
		{
			name: "should return perf profiler",
			tool: api.Perf,
			want: NewPerfProfiler(executil.NewCommander(), publish.NewPublisher()),
		},
		{
			name: "should return mock profiler",
			tool: api.FakeTool,
			want: NewMockProfiler(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Get(tt.tool)

			assert.Equalf(t, tt.want, got, "Get(%v)", tt.tool)
		})
	}
}
