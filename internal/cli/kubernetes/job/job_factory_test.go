package job

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/stretchr/testify/assert"
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
			name: "node with node dummy creator is instanced",
			args: args{
				lang: api.Node,
				tool: api.NodeDummy,
			},
			want: &dummyCreator{},
		},
		{
			name: "rust creator is instanced",
			args: args{
				lang: api.Rust,
			},
			want: &bpfCreator{},
		},
		{
			name: "rust with perf creator is instanced",
			args: args{
				lang: api.Rust,
				tool: api.Perf,
			},
			want: &perfCreator{},
		},
		{
			name: "clang creator is instanced",
			args: args{
				lang: api.Clang,
			},
			want: &bpfCreator{},
		},
		{
			name: "clang with perf creator is instanced",
			args: args{
				lang: api.Clang,
				tool: api.Perf,
			},
			want: &perfCreator{},
		},
		{
			name: "clang++ creator is instanced",
			args: args{
				lang: api.ClangPlusPlus,
			},
			want: &bpfCreator{},
		},
		{
			name: "clang++ with perf creator is instanced",
			args: args{
				lang: api.ClangPlusPlus,
				tool: api.Perf,
			},
			want: &perfCreator{},
		},
		{
			name: "fake creator is instanced",
			args: args{
				lang: api.FakeLang,
			},
			want: &fakeCreator{},
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
