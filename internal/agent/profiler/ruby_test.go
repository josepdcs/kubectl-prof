package profiler

import (
	"bytes"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

type MockRubyManager interface {
	RubyManager
	InvokeInvokedTimes() int
	PublishResultInvokedTimes() int
	WithInvokeError() MockRubyManager
}

type mockRubyManager struct {
	invokeInvokedTimes        int
	publishResultInvokedTimes int
	withInvokeError           bool
}

// NewMockRubyManager instances an empty MockRubyManager util for unit tests
func NewMockRubyManager() MockRubyManager {
	return &mockRubyManager{}
}

func (m *mockRubyManager) invoke(*job.ProfilingJob, string) (error, time.Duration) {
	m.invokeInvokedTimes++
	if m.withInvokeError {
		return errors.New("fake handleFlamegraph with error"), 0
	}
	fmt.Println("fake invoke")
	return nil, 0
}

func (m *mockRubyManager) publishResult(compressor.Type, string, api.OutputType) error {
	m.publishResultInvokedTimes++
	fmt.Println("fake publish result")
	return nil
}

func (m *mockRubyManager) InvokeInvokedTimes() int {
	return m.invokeInvokedTimes
}

func (m *mockRubyManager) PublishResultInvokedTimes() int {
	return m.publishResultInvokedTimes
}

func (m *mockRubyManager) WithInvokeError() MockRubyManager {
	m.withInvokeError = true
	return m
}

func TestRubyProfiler_SetUp(t *testing.T) {
	type fields struct {
		RubyProfiler *RubyProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, fields fields)
	}{
		{
			name: "should setup",
			given: func() (fields, args) {
				return fields{
						RubyProfiler: &RubyProfiler{
							RubyManager: NewMockRubyManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.RubyProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.RubyProfiler.targetPIDs)
			},
		},
		{
			name: "should setup with given PID",
			given: func() (fields, args) {
				return fields{
						RubyProfiler: &RubyProfiler{
							RubyManager: NewMockRubyManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							PID:              "PID_ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.RubyProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.RubyProfiler.targetPIDs)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						RubyProfiler: &RubyProfiler{
							RubyManager: NewMockRubyManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainerWithPIDResultError,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.RubyProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.RubyProfiler.targetPIDs)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err := tt.when(fields, args)

			// Then
			tt.then(t, err, fields)
		})
	}
}

func TestRubyProfiler_Invoke(t *testing.T) {
	type fields struct {
		RubyProfiler *RubyProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, time.Duration)
		then  func(t *testing.T, err error, fields fields)
	}{
		{
			name: "should invoke",
			given: func() (fields, args) {
				return fields{
						RubyProfiler: &RubyProfiler{
							RubyManager: NewMockRubyManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				fields.RubyProfiler.delay = 0
				fields.RubyProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.RubyProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.RubyProfiler.RubyManager.(MockRubyManager)
				assert.Nil(t, err)
				assert.Equal(t, 2, mock.InvokeInvokedTimes())
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				return fields{
						RubyProfiler: &RubyProfiler{
							RubyManager: NewMockRubyManager().WithInvokeError(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				fields.RubyProfiler.delay = 0
				fields.RubyProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.RubyProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.RubyProfiler.RubyManager.(MockRubyManager)
				require.Error(t, err)
				assert.Equal(t, 1, mock.InvokeInvokedTimes())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err, _ := tt.when(fields, args)

			// Then
			tt.then(t, err, fields)
		})
	}
}

func TestRubyProfiler_CleanUp(t *testing.T) {
	type fields struct {
		RubyProfiler *RubyProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, fields fields)
	}{
		{
			name: "should cleanup",
			given: func() (fields, args) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
						RubyProfiler: &RubyProfiler{
							RubyManager: NewMockRubyManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.RubyProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg"+
					compressor.GetExtensionFileByCompressor[compressor.Gzip])
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
				assert.Nil(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err := tt.when(fields, args)

			// Then
			tt.then(t, err, fields)
		})
	}
}

func Test_rubyManager_invoke(t *testing.T) {
	type fields struct {
		RubyProfiler *RubyProfiler
	}
	type args struct {
		job *job.ProfilingJob
		pid string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, time.Duration)
		then  func(t *testing.T, err error)
		after func()
	}{
		{
			name: "should invoke",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000.txt"), b.String())
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"), b.String())

				rbSpyCommander = executil.NewFakeCommander(exec.Command("ls", "/tmp"))
				return fields{
						RubyProfiler: NewRubyProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
							Tool:             api.Rbspy,
							Compressor:       compressor.None,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.RubyProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error) {
				assert.Nil(t, err)
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg")))
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000.txt"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"))
			},
		},
		{
			name: "should invoke fail when command fail",
			given: func() (fields, args) {
				rbSpyCommander = executil.NewFakeCommander(&exec.Cmd{})
				return fields{
						RubyProfiler: NewRubyProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Ruby,
							Tool:             api.Rbspy,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.RubyProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err, _ := tt.when(fields, args)

			// Then
			tt.then(t, err)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_rubyManager_publishResult(t *testing.T) {
	p := NewRubyProfiler()
	err := p.publishResult(compressor.Gzip, testdata.ResultTestDataDir()+"/flamegraph.svg", api.FlameGraph)
	assert.Nil(t, err)
}
