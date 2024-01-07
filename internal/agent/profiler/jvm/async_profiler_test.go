package jvm

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

type MockAsyncProfilerManager interface {
	AsyncProfilerManager
	RemoveTmpDirInvokedTimes() int
	LinkTmpDirToTargetTmpDirInvokedTimes() int
	CopyProfilerToTmpDirInvokedTimes() int
	PublishResultInvokedTimes() int
	WithRemoveTmpDirResultError() MockAsyncProfilerManager
	WithLinkTmpDirToTargetTmpDirResultError() MockAsyncProfilerManager
}

type mockAsyncProfilerManager struct {
	removeTmpDirInvokedTimes                int
	linkTmpDirToTargetTmpDirInvokedTimes    int
	copyProfilerToTmpDirInvokedTimes        int
	publishResultInvokedTimes               int
	withRemoveTmpDirResultError             bool
	withLinkTmpDirToTargetTmpDirResultError bool
}

// NewMockAsyncProfilerManager instances an empty MockAsyncProfilerManager util for unit tests
func NewMockAsyncProfilerManager() MockAsyncProfilerManager {
	return &mockAsyncProfilerManager{}
}

func (m *mockAsyncProfilerManager) removeTmpDir() error {
	m.removeTmpDirInvokedTimes++
	if m.withRemoveTmpDirResultError {
		return errors.New("fake removeTmpDir with error")
	}
	fmt.Println("fake removeTmpDir")
	return nil
}

func (m *mockAsyncProfilerManager) linkTmpDirToTargetTmpDir(string) error {
	m.linkTmpDirToTargetTmpDirInvokedTimes++
	if m.withLinkTmpDirToTargetTmpDirResultError {
		return errors.New("fake linkTmpDirToTargetTmpDir with error")
	}
	fmt.Println("fake linkTmpDirToTargetTmpDir")
	return nil
}

func (m *mockAsyncProfilerManager) copyProfilerToTmpDir() error {
	fmt.Println("fake copy")
	m.copyProfilerToTmpDirInvokedTimes++
	return nil
}

func (m *mockAsyncProfilerManager) publishResult(compressor.Type, string, api.OutputType) error {
	fmt.Println("fake publish result")
	m.publishResultInvokedTimes++
	return nil
}

func (m *mockAsyncProfilerManager) RemoveTmpDirInvokedTimes() int {
	return m.removeTmpDirInvokedTimes
}

func (m *mockAsyncProfilerManager) LinkTmpDirToTargetTmpDirInvokedTimes() int {
	return m.linkTmpDirToTargetTmpDirInvokedTimes
}

func (m *mockAsyncProfilerManager) CopyProfilerToTmpDirInvokedTimes() int {
	return m.copyProfilerToTmpDirInvokedTimes
}

func (m *mockAsyncProfilerManager) PublishResultInvokedTimes() int {
	return m.publishResultInvokedTimes
}

func (m *mockAsyncProfilerManager) WithRemoveTmpDirResultError() MockAsyncProfilerManager {
	m.withRemoveTmpDirResultError = true
	return m
}

func (m *mockAsyncProfilerManager) WithLinkTmpDirToTargetTmpDirResultError() MockAsyncProfilerManager {
	m.withLinkTmpDirToTargetTmpDirResultError = true
	return m
}

func TestNewAsyncProfiler(t *testing.T) {
	p := NewAsyncProfiler()
	assert.IsType(t, p, &AsyncProfiler{})
}

func TestAsyncProfiler_SetUp(t *testing.T) {
	type fields struct {
		AsyncProfiler *AsyncProfiler
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
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager(),
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.AsyncProfiler.AsyncProfilerManager.(MockAsyncProfilerManager)
				assert.Nil(t, err)
				assert.Equal(t, "PID_ContainerID", fields.AsyncProfiler.targetPID)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.CopyProfilerToTmpDirInvokedTimes())
			},
		},
		{
			name: "should setup when PID is given",
			given: func() (fields, args) {
				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager(),
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.AsyncProfiler.AsyncProfilerManager.(MockAsyncProfilerManager)
				assert.Nil(t, err)
				assert.Equal(t, "PID_ContainerID", fields.AsyncProfiler.targetPID)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.CopyProfilerToTmpDirInvokedTimes())
			},
		},
		{
			name: "should fail when getting target filesystem fail",
			given: func() (fields, args) {
				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: "other",
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.AsyncProfiler.AsyncProfilerManager.(MockAsyncProfilerManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.CopyProfilerToTmpDirInvokedTimes())
			},
		},
		{
			name: "should fail when removing tmp dir fail",
			given: func() (fields, args) {
				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager().WithRemoveTmpDirResultError(),
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.AsyncProfiler.AsyncProfilerManager.(MockAsyncProfilerManager)
				assert.NotNil(t, err)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.CopyProfilerToTmpDirInvokedTimes())
			},
		},
		{
			name: "should fail when link tmp dir to target tmp dir fail",
			given: func() (fields, args) {
				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager().WithLinkTmpDirToTargetTmpDirResultError(),
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.AsyncProfiler.AsyncProfilerManager.(MockAsyncProfilerManager)
				assert.NotNil(t, err)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.CopyProfilerToTmpDirInvokedTimes())
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						AsyncProfiler: &AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager(),
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
				return fields.AsyncProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.AsyncProfiler.AsyncProfilerManager.(MockAsyncProfilerManager)
				assert.NotNil(t, err)
				assert.Empty(t, fields.AsyncProfiler.targetPID)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.CopyProfilerToTmpDirInvokedTimes())
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

func TestAsyncProfiler_Invoke(t *testing.T) {
	type fields struct {
		AsyncProfiler AsyncProfiler
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
				commander := executil.NewFakeCommander()
				commander.Return(exec.Command("ls", common.TmpDir())).Return(&exec.Cmd{}).On("Command")
				asyncProfilerCommander = commander
				return fields{
						AsyncProfiler: AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager(),
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
				return fields.AsyncProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.AsyncProfiler.AsyncProfilerManager.(MockAsyncProfilerManager)
				assert.Nil(t, err)
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should publish result when raw output type",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.Return(exec.Command("ls", common.TmpDir())).On("Command")
				asyncProfilerCommander = commander
				return fields{
						AsyncProfiler: AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.Raw,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.AsyncProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.AsyncProfiler.AsyncProfilerManager.(MockAsyncProfilerManager)
				assert.Nil(t, err)
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when fail exec command",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.Return(&exec.Cmd{}).On("Command")
				asyncProfilerCommander = commander
				return fields{
						AsyncProfiler: AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager(),
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
				return fields.AsyncProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.AsyncProfiler.AsyncProfilerManager.(MockAsyncProfilerManager)
				assert.NotNil(t, err)
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

func TestAsyncProfiler_CleanUp(t *testing.T) {
	type fields struct {
		AsyncProfiler AsyncProfiler
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
			name: "should clean up",
			given: func() (fields, args) {
				_ = os.Mkdir(filepath.Join(common.TmpDir(), "async-profiler"), os.ModePerm)
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])

				commander := executil.NewFakeCommander()
				commander.Return(exec.Command("ls", common.TmpDir())).On("Command")
				asyncProfilerCommander = commander
				return fields{
						AsyncProfiler: AsyncProfiler{
							AsyncProfilerManager: NewMockAsyncProfilerManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							UID:              "UID",
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							Compressor:       compressor.Gzip,
							Tool:             api.AsyncProfiler,
							OutputType:       api.FlameGraph,
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.AsyncProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html"+
					compressor.GetExtensionFileByCompressor[compressor.Gzip])
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
				assert.False(t, file.Exists(filepath.Join(common.TmpDir(), "async-profiler")))
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

func Test_asyncProfilerManager_copyProfilerToTmpDir(t *testing.T) {
	commander := executil.NewFakeCommander()
	commander.Return(exec.Command("ls", common.TmpDir())).On("Command")
	asyncProfilerCommander = commander
	a := NewAsyncProfiler()
	assert.Nil(t, a.copyProfilerToTmpDir())
}

func Test_asyncProfilerManager_publishResult(t *testing.T) {
	p := NewAsyncProfiler()
	err := p.publishResult(compressor.Gzip, testdata.ResultTestDataDir()+"/flamegraph.svg", api.FlameGraph)
	assert.Nil(t, err)
}
