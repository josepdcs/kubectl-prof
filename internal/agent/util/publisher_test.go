package util

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestPublish(t *testing.T) {
	type args struct {
		c         compressor.Type
		file      string
		eventType api.OutputType
	}
	tests := []struct {
		name       string
		args       args
		wantErrMsg string
		afterEach  func()
	}{
		{
			name: "should publish event",
			args: args{
				c:         compressor.Gzip,
				file:      testdata.ResultTestDataDir() + "/flamegraph.svg",
				eventType: api.FlameGraph,
			},
			afterEach: func() {
				_ = os.Remove(testdata.ResultTestDataDir() + "/flamegraph.svg.gz")
			},
		},
		{
			name: "should fail if file not exists",
			args: args{
				c:         compressor.Gzip,
				file:      testdata.ResultTestDataDir() + "/other",
				eventType: api.FlameGraph,
			},
			wantErrMsg: "no such file or directory",
		},
		{
			name: "should fail if compressor not exists",
			args: args{
				c:         compressor.Type("other"),
				file:      testdata.ResultTestDataDir() + "/flamegraph.svg",
				eventType: api.FlameGraph,
			},
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

			if tt.afterEach != nil {
				tt.afterEach()
			}

		})
	}
}
