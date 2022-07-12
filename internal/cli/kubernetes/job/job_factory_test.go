package job

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGet(t *testing.T) {
	type args struct {
		lang api.ProgrammingLanguage
		tool api.ProfilingTool
	}
	tests := []struct {
		name            string
		args            args
		want            Creator
		containedErrMsg string
	}{
		{
			name: "jvm creator is instanced",
			args: args{
				lang: api.Java,
			},
			want: &jvmCreator{},
		},
		{
			name: "go creator is instanced",
			args: args{
				lang: api.Go,
			},
			want: &bpfCreator{},
		},
		{
			name: "python creator is instanced",
			args: args{
				lang: api.Python,
			},
			want: &pythonCreator{},
		},
		{
			name: "ruby creator is instanced",
			args: args{
				lang: api.Ruby,
			},
			want: &rubyCreator{},
		},
		{
			name: "node creator is instanced",
			args: args{
				lang: api.Node,
			},
			want: &bpfCreator{},
		},
		{
			name: "node with perf creator is instanced",
			args: args{
				lang: api.Node,
				tool: api.Perf,
			},
			want: &perfCreator{},
		},
		{
			name: "creator not found",
			args: args{
				lang: api.ProgrammingLanguage("other"),
			},
			containedErrMsg: "ot language without job creator",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.lang, tt.args.tool)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			} else {
				assert.IsType(t, tt.want, got)
			}
		})
	}
}
