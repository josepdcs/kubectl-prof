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
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
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
	WithInvokeError() MockBpfManager
}

type mockBpfManager struct {
	invokeInvokedTimes                int
	handleProfilingResultInvokedTimes int
	publishResultInvokedTimes         int
	withHandleProfilingResultError    bool
	withInvokeError                   bool
}

// NewMockBpfManager instances an empty MockBpfManager util for unit tests
func NewMockBpfManager() MockBpfManager {
	return &mockBpfManager{}
}

func (m *mockBpfManager) invoke(*job.ProfilingJob, string) (error, time.Duration) {
	m.invokeInvokedTimes++
	if m.withInvokeError {
		return errors.New("fake invoke with error"), 0
	}
	fmt.Println("fake invoke")
	return nil, 0
}

func (m *mockBpfManager) handleFlamegraph(*job.ProfilingJob, flamegraph.FrameGrapher, string, string) error {
	m.handleProfilingResultInvokedTimes++
	if m.withHandleProfilingResultError {
		return errors.New("fake handleFlamegraph with error")
	}
	fmt.Println("fake handleFlamegraph")
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

func (m *mockBpfManager) WithInvokeError() MockBpfManager {
	m.withInvokeError = true
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
			name: "should invoke",
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
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				return fields{
						BpfProfiler: &BpfProfiler{
							BpfManager: NewMockBpfManager().WithInvokeError(),
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
							Tool:             api.Bpf,
							Compressor:       compressor.None,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.BpfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error) {
				assert.Nil(t, err)
				//assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg")))
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw-1000.txt"))
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"))
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
							Language:         api.Go,
							Tool:             api.Bpf,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.BpfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "should invoke return nil when fail handle flamegraph",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				bpfCommander = executil.NewFakeCommander(exec.Command("ls", "/tmp"))
				return fields{
						BpfProfiler: NewBpfProfiler(),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Go,
							Tool:             api.Bpf,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.BpfProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, err error) {
				require.NoError(t, err)
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

func Test_bpfManager_handleFlamegraph(t *testing.T) {
	type fields struct {
		BpfProfiler *BpfProfiler
	}
	type args struct {
		job            *job.ProfilingJob
		flameGrapher   flamegraph.FrameGrapher
		fileName       string
		resultFileName string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, flameGrapher flamegraph.FrameGrapher)
		after func()
	}{
		{
			name: "should handle flamegraph",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("testtesttesttesttesttesttesttesttesttesttesttesttest"))
				_ = os.WriteFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"), b.Bytes(), 0644)
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
						flameGrapher:   flamegraph.NewFlameGrapherFake(),
						fileName:       filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
						resultFileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
			},
			then: func(t *testing.T, err error, flameGrapher flamegraph.FrameGrapher) {
				assert.True(t, flameGrapher.(*flamegraph.FlameGrapherFake).StackSamplesToFlameGraphInvoked)
				assert.Nil(t, err)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"))
			},
		},
		{
			name: "should fail handle flamegraph profiler result",
			given: func() (fields, args) {
				var b bytes.Buffer
				b.Write([]byte("testtesttesttesttesttesttesttesttesttesttesttesttest"))
				_ = os.WriteFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"), b.Bytes(), 0644)
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
						flameGrapher:   flamegraph.NewFlameGrapherFakeWithError(),
						fileName:       filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
						resultFileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
			},
			then: func(t *testing.T, err error, flameGrapher flamegraph.FrameGrapher) {
				assert.True(t, flameGrapher.(*flamegraph.FlameGrapherFakeWithError).StackSamplesToFlameGraphInvoked)
				assert.EqualError(t, err, "could not convert raw format to flamegraph: StackSamplesToFlameGraph with error")
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"))
			},
		},
		{
			name: "should fail handle flamegraph profiler result when no stacks found",
			given: func() (fields, args) {
				var b bytes.Buffer
				_ = os.WriteFile(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"), b.Bytes(), 0644)
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
						flameGrapher:   flamegraph.NewFlameGrapherFake(),
						fileName:       filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"),
						resultFileName: filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.svg"),
					}
			},
			when: func(fields fields, args args) error {
				return fields.BpfProfiler.handleFlamegraph(args.job, args.flameGrapher, args.fileName, args.resultFileName)
			},
			then: func(t *testing.T, err error, flameGrapher flamegraph.FrameGrapher) {
				assert.False(t, flameGrapher.(*flamegraph.FlameGrapherFake).StackSamplesToFlameGraphInvoked)
				assert.EqualError(t, err, "unable to generate flamegraph: no stacks found (maybe due low cpu load)")
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"raw.txt"))
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
			tt.then(t, err, args.flameGrapher)

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
