package utils

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestPublish(t *testing.T) {
	type args struct {
		c         api.Compressor
		file      string
		eventType api.EventType
	}
	tests := []struct {
		name       string
		args       args
		wantErrMsg string
		clean      func()
	}{
		{
			name: "should publish event",
			args: args{
				c:         api.Gzip,
				file:      test.ResultTestDataDir() + "/flamegraph.svg",
				eventType: api.FlameGraph,
			},
			clean: func() {
				_ = os.Remove(test.ResultTestDataDir() + "/flamegraph.svg.gz")
			},
		},
		{
			name: "should fail if file not exists",
			args: args{
				c:         api.Gzip,
				file:      test.ResultTestDataDir() + "/other",
				eventType: api.FlameGraph,
			},
			clean:      func() {},
			wantErrMsg: "no such file or directory",
		},
		{
			name: "should fail if compressor not exists",
			args: args{
				c:         api.Compressor("other"),
				file:      test.ResultTestDataDir() + "/flamegraph.svg",
				eventType: api.FlameGraph,
			},
			clean:      func() {},
			wantErrMsg: "could not find compressor for other",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Publish(tt.args.c, tt.args.file, tt.args.eventType)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				require.NoError(t, err)
			}

			tt.clean()

		})
	}
}
