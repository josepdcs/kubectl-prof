package profiler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_addProcessPIDLegend(t *testing.T) {
	type args struct {
		input string
		pid   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should add process pid legend",
			args: args{
				input: "test",
				pid:   "123",
			},
			want: "process: 123;test\n",
		},
		{
			name: "should return input if input is empty",
			args: args{
				input: "",
				pid:   "123",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addProcessPIDLegend(tt.args.input, tt.args.pid)
			assert.Equalf(t, tt.want, got, "addProcessPIDLegend(%v, %v)", tt.args.input, tt.args.pid)
		})
	}
}
