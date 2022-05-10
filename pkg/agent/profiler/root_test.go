package profiler

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGet(t *testing.T) {
	type args struct {
		lang api.ProgrammingLanguage
		tool api.ProfilingTool
	}
	tests := []struct {
		name       string
		args       args
		want       Profiler
		wantErrMsg string
	}{
		{
			name: "should return jvm profiler for async profiler tool",
			args: args{
				lang: api.Java,
			},
			want: NewJvmProfiler(),
		},
		{
			name: "should return bpf profiler when go lang",
			args: args{
				lang: api.Go,
			},
			want: NewBpfProfiler(),
		},
		{
			name: "should return python profiler",
			args: args{
				lang: api.Python,
			},
			want: NewPythonProfiler(),
		},
		{
			name: "should return ruby profiler",
			args: args{
				lang: api.Ruby,
			},
			want: NewRubyProfiler(),
		},
		{
			name: "should return perf profiler when node use perf tool",
			args: args{
				lang: api.Node,
				tool: api.Perf,
			},
			want: NewPerfProfiler(),
		},
		{
			name: "should return bpf profiler when node use bpf tool",
			args: args{
				lang: api.Node,
				tool: api.Bpf,
			},
			want: NewBpfProfiler(),
		},
		{
			name: "should fail when lang not found",
			args: args{
				lang: api.ProgrammingLanguage("other"),
			},
			wantErrMsg: "could not find profiler for language",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.lang, tt.args.tool)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				require.NoError(t, err)
			}

			assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.lang, tt.args.tool)
		})
	}
}
