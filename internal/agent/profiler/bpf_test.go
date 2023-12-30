package profiler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

type MockBpfManager interface {
	BpfManager
	InvokeInvokedTimes() int
	HandleProfilingResultInvokedTimes() int
	PublishResultInvokedTimes() int
	WithHandleProfilingResultError() MockBpfManager
}

type mockBpfManager struct {
	invokeInvokedTimes                int
	handleProfilingResultInvokedTimes int
	publishResultInvokedTimes         int
	withHandleProfilingResultError    bool
}

// NewMockBpfManager instances an empty MockBpfManager util for unit tests
func NewMockBpfManager() MockBpfManager {
	return &mockBpfManager{}
}

func (m *mockBpfManager) invoke(job *job.ProfilingJob, pid string) (error, string, time.Duration) {
	m.invokeInvokedTimes++
	fmt.Println("fake invoke")
	return nil, "", 0
}

func (m *mockBpfManager) handleProfilingResult(*job.ProfilingJob, flamegraph.FrameGrapher, string) error {
	m.handleProfilingResultInvokedTimes++
	if m.withHandleProfilingResultError {
		return errors.New("fake handleProfilingResult with error")
	}
	fmt.Println("fake handleProfilingResult")
	return nil
}

func (m *mockBpfManager) publishResult(compressor.Type, string, api.OutputType) error {
	m.publishResultInvokedTimes++
	fmt.Println("fake publish result")
	return nil
}

func (m *mockBpfManager) InvokeInvokedTimes() int {
	return m.invokeInvokedTimes
}

func (m *mockBpfManager) HandleProfilingResultInvokedTimes() int {
	return m.handleProfilingResultInvokedTimes
}

func (m *mockBpfManager) PublishResultInvokedTimes() int {
	return m.publishResultInvokedTimes
}

func (m *mockBpfManager) WithHandleProfilingResultError() MockBpfManager {
	m.withHandleProfilingResultError = true
	return m
}

func TestBpfProfiler_SetUp(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
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
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
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
				return fields.BpfProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.BpfProfiler.targetPIDs)
			},
		},
		{
			name: "should setup with given PID",
			given: func() (fields, args) {
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
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
				return fields.BpfProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.BpfProfiler.targetPIDs)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
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
				return fields.BpfProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.BpfProfiler.targetPIDs)
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

func TestBpfProfiler_Invoke(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
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
			name: "should publish result",
			given: func() (fields, args) {
				bpfCommander = executil.NewFakeCommander(exec.Command("ls", "/tmp"))
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
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
				fields.BpfProfiler.delay = 0
				fields.BpfProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.BpfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.BpfProfiler.BpfManager.(MockBpfManager)
				assert.Nil(t, err)
				assert.Equal(t, 2, mock.InvokeInvokedTimes())
				assert.Equal(t, 1, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when handle profiling result fail",
			given: func() (fields, args) {
				bpfCommander = executil.NewFakeCommander(exec.Command("ls", "/tmp"))
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager().WithHandleProfilingResultError(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				fields.BpfProfiler.delay = 0
				fields.BpfProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.BpfProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.BpfProfiler.BpfManager.(MockBpfManager)
				assert.NotNil(t, err)
				assert.Equal(t, 2, mock.InvokeInvokedTimes())
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
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

func TestBpfProfiler_CleanUp(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
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
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager(),
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
				return fields.BpfProfiler.CleanUp(args.job)
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

func Test_bpfManager_invoke(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
	}
	type args struct {
		job *job.ProfilingJob
		pid string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, string, time.Duration)
		then  func(t *testing.T, err error, fileName string)
		after func()
	}{
		{
			name: "should invoke",
			given: func() (fields, args) {
				bpfCommander = executil.NewFakeCommander(exec.Command("ls", "/tmp"))
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, string, time.Duration) {
				return fields.BpfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error, fileName string) {
				assert.Nil(t, err)
				assert.Equal(t, filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.svg.1000"), fileName)
			},
		},
		{
			name: "should invoke fail when command fail",
			given: func() (fields, args) {
				bpfCommander = executil.NewFakeCommander(&exec.Cmd{})
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, string, time.Duration) {
				return fields.BpfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error, fileName string) {
				require.Error(t, err)
				assert.Equal(t, "", fileName)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err, fileName, _ := tt.when(fields, args)

			// Then
			tt.then(t, err, fileName)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_bpfManager_handleProfilingResult(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
	}
	type args struct {
		job          *job.ProfilingJob
		flameGrapher flamegraph.FrameGrapher
		fileName     string
		out          bytes.Buffer
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, fields fields)
		after func()
	}{
		{
			name: "should handle flamegraph profiler result",
			given: func() (fields, args) {
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.FakeLang,
						},
						flameGrapher: flamegraph.NewFlameGrapherFake(),
						fileName:     filepath.Join(testdata.ResultTestDataDir(), "raw.txt"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
			},
		},
		{
			name: "should fail handle flamegraph profiler result",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("test"))
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         "other",
						},
						flameGrapher: flamegraph.NewFlameGrapherFakeWithError(),
						fileName:     filepath.Join(testdata.ResultTestDataDir(), "raw.txt"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.EqualError(t, err, "could not convert raw format to flamegraph: StackSamplesToFlameGraph with error")
			},
		},
		{
			name: "should fail handle flamegraph profiler result when no stacks found",
			given: func() (fields, args) {
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         "other",
						},
						flameGrapher: flamegraph.NewFlameGrapherFake(),
						fileName:     filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleProfilingResult(args.job, args.flameGrapher, args.fileName)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.EqualError(t, err, "unable to generate flamegraph: no stacks found (maybe due low cpu load)")
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

			if tt.after != nil {
				tt.after()
			}
		})
	}
}

func Test_bpfManager_bccProfilerCommand(t *testing.T) {
	// Given
	j := &job.ProfilingJob{
		Interval: 10 * time.Second,
	}

	// When
	result := bccProfilerCommand(j, "1000")

	// Then
	assert.NotNil(t, result)
}

func Test_bpfManager_publishResult(t *testing.T) {
	p := NewBpfProfiler()
	err := p.publishResult(compressor.Gzip, testdata.ResultTestDataDir()+"/flamegraph.svg", api.FlameGraph)
	assert.Nil(t, err)
}
