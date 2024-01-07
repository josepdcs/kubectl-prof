package jvm

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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

type MockJcmdManager interface {
	JcmdManager
	RemoveTmpDirInvokedTimes() int
	LinkTmpDirToTargetTmpDirInvokedTimes() int
	CopyJfrSettingsToTmpDirInvokedTimes() int
	HandleProfilingResultInvokedTimes() int
	PublishResultInvokedTimes() int
	HandleJcmdRecordingTimes() int
	WithHandleProfilingResultError() MockJcmdManager
	WithRemoveTmpDirResultError() MockJcmdManager
	WithLinkTmpDirToTargetTmpDirResultError() MockJcmdManager
}

type mockJcmdManager struct {
	removeTmpDirInvokedTimes                int
	linkTmpDirToTargetTmpDirInvokedTimes    int
	copyJfrSettingsToTmpDirInvokedTimes     int
	handleProfilingResultInvokedTimes       int
	publishResultInvokedTimes               int
	handleJcmdRecordingTimes                int
	withHandleProfilingResultError          bool
	withRemoveTmpDirResultError             bool
	withLinkTmpDirToTargetTmpDirResultError bool
}

// NewMockJcmdManager instances an empty MockJcmdManager util for unit tests
func NewMockJcmdManager() MockJcmdManager {
	return &mockJcmdManager{}
}

func (m *mockJcmdManager) removeTmpDir() error {
	m.removeTmpDirInvokedTimes++
	if m.withRemoveTmpDirResultError {
		return errors.New("fake removeTmpDir with error")
	}
	fmt.Println("fake removeTmpDir")
	return nil
}

func (m *mockJcmdManager) linkTmpDirToTargetTmpDir(string) error {
	m.linkTmpDirToTargetTmpDirInvokedTimes++
	if m.withLinkTmpDirToTargetTmpDirResultError {
		return errors.New("fake linkTmpDirToTargetTmpDir with error")
	}
	fmt.Println("fake linkTmpDirToTargetTmpDir")
	return nil
}

func (m *mockJcmdManager) copyJfrSettingsToTmpDir() error {
	fmt.Println("fake copyJfrSettingsToTmpDir")
	m.copyJfrSettingsToTmpDirInvokedTimes++
	return nil
}

func (m *mockJcmdManager) handleProfilingResult(*job.ProfilingJob, string, bytes.Buffer, string) error {
	m.handleProfilingResultInvokedTimes++
	if m.withHandleProfilingResultError {
		return errors.New("fake handleProfilingResult with error")
	}
	fmt.Println("fake handleProfilingResult")
	return nil
}

func (m *mockJcmdManager) handleJcmdRecording(string, string) {
	fmt.Println("fake handleJcmdRecording")
}

func (m *mockJcmdManager) publishResult(compressor.Type, string, api.OutputType, string) error {
	fmt.Println("fake publish result")
	m.publishResultInvokedTimes++
	return nil
}

func (m *mockJcmdManager) RemoveTmpDirInvokedTimes() int {
	return m.removeTmpDirInvokedTimes
}

func (m *mockJcmdManager) LinkTmpDirToTargetTmpDirInvokedTimes() int {
	return m.linkTmpDirToTargetTmpDirInvokedTimes
}

func (m *mockJcmdManager) CopyJfrSettingsToTmpDirInvokedTimes() int {
	return m.copyJfrSettingsToTmpDirInvokedTimes
}

func (m *mockJcmdManager) HandleProfilingResultInvokedTimes() int {
	return m.handleProfilingResultInvokedTimes
}

func (m *mockJcmdManager) PublishResultInvokedTimes() int {
	return m.publishResultInvokedTimes
}

func (m *mockJcmdManager) HandleJcmdRecordingTimes() int {
	return m.handleJcmdRecordingTimes
}

func (m *mockJcmdManager) WithHandleProfilingResultError() MockJcmdManager {
	m.withHandleProfilingResultError = true
	return m
}

func (m *mockJcmdManager) WithRemoveTmpDirResultError() MockJcmdManager {
	m.withRemoveTmpDirResultError = true
	return m
}

func (m *mockJcmdManager) WithLinkTmpDirToTargetTmpDirResultError() MockJcmdManager {
	m.withLinkTmpDirToTargetTmpDirResultError = true
	return m
}

func TestJcmdProfiler_SetUp(t *testing.T) {
	type fields struct {
		JcmdProfiler *JcmdProfiler
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
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
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
				return fields.JcmdProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.Nil(t, err)
				assert.Equal(t, "PID_ContainerID", fields.JcmdProfiler.targetPID)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.CopyJfrSettingsToTmpDirInvokedTimes())
			},
		},
		{
			name: "should setup when PID is given",
			given: func() (fields, args) {
				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
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
				return fields.JcmdProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.Nil(t, err)
				assert.Equal(t, "PID_ContainerID", fields.JcmdProfiler.targetPID)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.CopyJfrSettingsToTmpDirInvokedTimes())
			},
		},
		{
			name: "should fail when getting target filesystem fail",
			given: func() (fields, args) {
				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.Crio,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.JcmdProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.NotNil(t, err)
				assert.Empty(t, fields.JcmdProfiler.targetPID)
				assert.Equal(t, 0, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.CopyJfrSettingsToTmpDirInvokedTimes())
			},
		},
		{
			name: "should fail when removing tmp dir fail",
			given: func() (fields, args) {
				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: NewMockJcmdManager().WithRemoveTmpDirResultError(),
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
				return fields.JcmdProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.NotNil(t, err)
				assert.Empty(t, fields.JcmdProfiler.targetPID)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.CopyJfrSettingsToTmpDirInvokedTimes())
			},
		},
		{
			name: "should fail when link tmp dir to target tmp dir fail",
			given: func() (fields, args) {
				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: NewMockJcmdManager().WithLinkTmpDirToTargetTmpDirResultError(),
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
				return fields.JcmdProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.NotNil(t, err)
				assert.Empty(t, fields.JcmdProfiler.targetPID)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.CopyJfrSettingsToTmpDirInvokedTimes())
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
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
				return fields.JcmdProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.NotNil(t, err)
				assert.Empty(t, fields.JcmdProfiler.targetPID)
				assert.Equal(t, 1, mock.RemoveTmpDirInvokedTimes())
				assert.Equal(t, 1, mock.LinkTmpDirToTargetTmpDirInvokedTimes())
				assert.Equal(t, 0, mock.CopyJfrSettingsToTmpDirInvokedTimes())
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

func TestJcmdProfiler_Invoke(t *testing.T) {
	type fields struct {
		JcmdProfiler JcmdProfiler
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
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				jcmdCommander = commander
				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
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
				return fields.JcmdProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.Nil(t, err)
				assert.Equal(t, 1, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should publish result when ThreadDump output type",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				jcmdCommander = commander
				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         time.Duration(1) * time.Second,
							Interval:         time.Duration(1) * time.Second,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.JcmdProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.Nil(t, err)
				assert.Equal(t, 1, mock.HandleProfilingResultInvokedTimes())
				assert.Equal(t, 1, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when fail exec command",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.On("Command").Return(&exec.Cmd{})
				jcmdCommander = commander
				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
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
				return fields.JcmdProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when handle profiling result fail",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				jcmdCommander = commander
				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: NewMockJcmdManager().WithHandleProfilingResultError(),
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
				return fields.JcmdProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when profile fail for ThreadDump output type",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				jcmdCommander = commander
				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: NewMockJcmdManager().WithHandleProfilingResultError(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         time.Duration(2) * time.Second,
							Interval:         time.Duration(1) * time.Second,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.JcmdProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
				assert.NotNil(t, err)
				assert.Equal(t, 0, mock.PublishResultInvokedTimes())
			},
		},
		{
			name: "should fail when profile fail for ThreadDump output type",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				jcmdCommander = commander
				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: NewMockJcmdManager().WithHandleProfilingResultError(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         time.Duration(2) * time.Second,
							Interval:         time.Duration(1) * time.Second,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.ThreadDump,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.JcmdProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				mock := fields.JcmdProfiler.JcmdManager.(MockJcmdManager)
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

func TestJcmdProfiler_CleanUp(t *testing.T) {
	type fields struct {
		JcmdProfiler      JcmdProfiler
		stopJcmdRecording chan bool
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
				jcmdStopCommand = func(job *job.ProfilingJob, pid string) *exec.Cmd {
					return exec.Command("ls", common.TmpDir())
				}
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:            0,
							ContainerRuntime:    api.FakeContainer,
							ContainerID:         "ContainerID",
							Compressor:          compressor.Gzip,
							FileName:            "flamegraph.html",
							AdditionalArguments: map[string]string{"settings": "contprof"},
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.JcmdProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html"+
					compressor.GetExtensionFileByCompressor[compressor.Gzip])
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
				assert.Nil(t, err)
			},
		},
		{
			name: "should clean up when stopJcmdRecording",
			given: func() (fields, args) {
				stopJcmdRecording = make(chan bool, 1)
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
						},
						stopJcmdRecording: stopJcmdRecording,
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							Compressor:       compressor.Gzip,
							FileName:         "flamegraph.html",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.JcmdProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"-flamegraph.html")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"-flamegraph.html"+
					compressor.GetExtensionFileByCompressor[compressor.Gzip])
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
				assert.Nil(t, err)
				assert.True(t, <-fields.stopJcmdRecording)
			},
		},
		{
			name: "should clean no jcmd command stop",
			given: func() (fields, args) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph.html")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: NewMockJcmdManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							Compressor:       compressor.Gzip,
							FileName:         "flamegraph.html",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.JcmdProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"-flamegraph.html")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"-flamegraph.html"+
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

func Test_jcmdUtil_publishResult(t *testing.T) {
	type fields struct {
		JcmdProfiler JcmdProfiler
	}
	type args struct {
		c          compressor.Type
		fileName   string
		outputType api.OutputType
	}
	tests := []struct {
		name      string
		given     func() (fields, args)
		when      func(fields, args) error
		then      func(t *testing.T, err error)
		afterEach func()
	}{
		{
			name: "should publish heap dump",
			given: func() (fields, args) {
				cmd := exec.Command("cp", filepath.Join(testdata.ResultTestDataDir(), "heapdump.hprof"), common.TmpDir())
				_ = cmd.Run()
				return fields{
						JcmdProfiler: *NewJcmdProfiler(),
					}, args{
						c:          compressor.None,
						fileName:   filepath.Join(common.TmpDir(), "heapdump.hprof"),
						outputType: api.HeapDump,
					}
			},
			when: func(fields fields, args args) error {
				return fields.JcmdProfiler.JcmdManager.publishResult(args.c, args.fileName, args.outputType, "10M")
			},
			then: func(t *testing.T, err error) {
				assert.Nil(t, err)
				assert.False(t, file.Exists(filepath.Join(common.TmpDir(), "heapdump.hprof")))
				assert.False(t, file.Exists(filepath.Join(common.TmpDir(), "heapdump.hprof.gz")))
			},
			afterEach: func() {
				file.RemoveAll(common.TmpDir(), "heapdump.hprof.gz.*")
			},
		},
		{
			name: "should publish other output type",
			given: func() (fields, args) {
				return fields{
						JcmdProfiler: *NewJcmdProfiler(),
					}, args{
						c:          compressor.None,
						fileName:   testdata.ResultTestDataDir() + "/flamegraph.svg",
						outputType: api.FlameGraph,
					}
			},
			when: func(fields fields, args args) error {
				return fields.JcmdProfiler.JcmdManager.publishResult(args.c, args.fileName, args.outputType, "")
			},
			then: func(t *testing.T, err error) {
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
			tt.then(t, err)

			if tt.afterEach != nil {
				tt.afterEach()
			}
		})
	}
}
