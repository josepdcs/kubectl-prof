package util

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
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
				file:      filepath.Join(testdata.ResultTestDataDir(), "flamegraph.svg"),
				eventType: api.FlameGraph,
			},
			afterEach: func() {
				_ = os.Remove(filepath.Join(testdata.ResultTestDataDir(), "flamegraph.svg.gz"))
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
				file:      filepath.Join(testdata.ResultTestDataDir(), "flamegraph.svg"),
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

func TestPublishWithNativeGzip(t *testing.T) {
	type args struct {
		file      string
		chunkSize string
		eventType api.OutputType
	}
	tests := []struct {
		name       string
		beforeEach func()
		args       args
		wantErrMsg string
		afterEach  func()
	}{
		{
			name: "should publish event",
			beforeEach: func() {
				cmd := exec.Command("cp", filepath.Join(testdata.ResultTestDataDir(), "flamegraph.svg"), common.TmpDir())
				_ = cmd.Run()
			},
			args: args{
				file:      filepath.Join(common.TmpDir(), "flamegraph.svg"),
				chunkSize: "1K",
				eventType: api.FlameGraph,
			},
			afterEach: func() {
				file.RemoveAll(common.TmpDir(), "flamegraph.svg.gz.*")
			},
		},
		{
			name: "should fail if file not exists",
			args: args{
				file:      testdata.ResultTestDataDir() + "/other",
				eventType: api.FlameGraph,
			},
			wantErrMsg: "does not exist",
		},
		{
			name: "should fail if chunk size not given",
			beforeEach: func() {
				cmd := exec.Command("cp", filepath.Join(testdata.ResultTestDataDir(), "flamegraph.svg"), common.TmpDir())
				_ = cmd.Run()
			},
			args: args{
				file:      filepath.Join(common.TmpDir(), "flamegraph.svg"),
				eventType: api.FlameGraph,
			},
			wantErrMsg: "chunk size is mandatory",
			afterEach: func() {
				_ = os.Remove(filepath.Join(common.TmpDir(), "flamegraph.svg.gz"))
			},
		},
		{
			name: "should fail if split command fails",
			beforeEach: func() {
				cmd := exec.Command("cp", filepath.Join(testdata.ResultTestDataDir(), "flamegraph.svg"), common.TmpDir())
				_ = cmd.Run()
			},
			args: args{
				file:      filepath.Join(common.TmpDir(), "flamegraph.svg"),
				chunkSize: "XX",
				eventType: api.FlameGraph,
			},
			wantErrMsg: "split failed on file",
			afterEach: func() {
				_ = os.Remove(filepath.Join(common.TmpDir(), "flamegraph.svg.gz"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeEach != nil {
				tt.beforeEach()
			}

			err := PublishWithNativeGzipAndSplit(tt.args.file, tt.args.chunkSize, tt.args.eventType)

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
