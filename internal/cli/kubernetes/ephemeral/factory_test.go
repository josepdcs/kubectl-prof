package ephemeral

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	type args struct {
		lang api.ProgrammingLanguage
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
			name: "node creator is instanced",
			args: args{
				lang: api.Node,
			},
			want: &bpfCreator{},
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
			got, err := Get(tt.args.lang)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			} else {
				assert.IsType(t, tt.want, got)
			}
		})
	}
}
