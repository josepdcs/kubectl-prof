package flamegraph

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name string
		job  *job.ProfilingJob
		want FrameGrapher
	}{
		{
			name: "should return flame grapher for python",
			job:  &job.ProfilingJob{Language: api.Python, Event: api.Cpu},
			want: NewFlameGrapherScript(
				WithTitle("PYTHON - CPU Flamegraph")),
		},
		{
			name: "should return flame grapher for golang",
			job:  &job.ProfilingJob{Language: api.Go, Event: api.Cpu},
			want: NewFlameGrapherScript(
				WithTitle("GO - CPU Flamegraph")),
		},
		{
			name: "should return flame grapher for node",
			job:  &job.ProfilingJob{Language: api.Node, Event: api.Cpu},
			want: NewFlameGrapherScript(
				WithTitle("NODE - CPU Flamegraph"),
				WithColors("js")),
		},
		{
			name: "should return flame grapher for fake language",
			job:  &job.ProfilingJob{Language: api.FakeLang},
			want: NewFlameGrapherFake(),
		},
		{
			name: "should return default flame grapher",
			job:  &job.ProfilingJob{Language: ""},
			want: NewFlameGrapherFakeWithError(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Get(tt.job), "Get(%v)", tt.job)
		})
	}
}
