package profiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDotnetProfiler_SetUp(t *testing.T) {
	type fields struct {
		DotnetProfiler *DotnetProfiler
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
						DotnetProfiler: &DotnetProfiler{
							DotnetManager: newMockDotnetManager(),
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
				return fields.DotnetProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.DotnetProfiler.targetPIDs)
			},
		},
		{
			name: "should setup with given PID",
			given: func() (fields, args) {
				return fields{
						DotnetProfiler: &DotnetProfiler{
							DotnetManager: newMockDotnetManager(),
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
				return fields.DotnetProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.DotnetProfiler.targetPIDs)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				return fields{
						DotnetProfiler: &DotnetProfiler{
							DotnetManager: newMockDotnetManager(),
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
				return fields.DotnetProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.DotnetProfiler.targetPIDs)
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

func TestDotnetProfiler_Invoke(t *testing.T) {
	type fields struct {
		DotnetProfiler *DotnetProfiler
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
				dotnetManager := newMockDotnetManager()
				dotnetManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, time.Duration(0)).
					Twice()

				return fields{
						DotnetProfiler: &DotnetProfiler{
							DotnetManager: dotnetManager,
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.SpeedScope,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				fields.DotnetProfiler.delay = 0
				fields.DotnetProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.DotnetProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				fields.DotnetProfiler.DotnetManager.(*mockDotnetManager).AssertNumberOfCalls(t, "invoke", 2)
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				dotnetManager := newMockDotnetManager()
				dotnetManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
					Return(errors.New("fake invoke error"), time.Duration(0)).
					Once()

				return fields{
						DotnetProfiler: &DotnetProfiler{
							DotnetManager: dotnetManager,
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.SpeedScope,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				fields.DotnetProfiler.delay = 0
				fields.DotnetProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.DotnetProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
				assert.EqualError(t, err, "fake invoke error")
				fields.DotnetProfiler.DotnetManager.(*mockDotnetManager).AssertNumberOfCalls(t, "invoke", 1)
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

func TestDotnetProfiler_CleanUp(t *testing.T) {
	type fields struct {
		DotnetProfiler *DotnetProfiler
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
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"speedscope-1000-1.json")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])
				return fields{
						DotnetProfiler: &DotnetProfiler{
							DotnetManager: newMockDotnetManager(),
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
				return fields.DotnetProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"speedscope-1000-1.json")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"speedscope-1000-1.json"+
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

func Test_dotnetManager_invoke(t *testing.T) {
	type fields struct {
		DotnetProfiler *DotnetProfiler
	}
	type args struct {
		job *job.ProfilingJob
		pid string
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, time.Duration)
		then  func(t *testing.T, fields fields, err error)
		after func()
	}{
		{
			name: "should invoke dotnet-trace with speedscope output",
			given: func() (fields, args) {
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						DotnetProfiler: NewDotnetProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.SpeedScope,
							Language:         api.DotNet,
							Tool:             api.DotnetTrace,
							Compressor:       compressor.None,
							Interval:         30 * time.Second,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.DotnetProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				assert.True(t, fields.DotnetProfiler.DotnetManager.(*dotnetManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
		},
		{
			name: "should invoke dotnet-trace with raw (nettrace) output",
			given: func() (fields, args) {
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						DotnetProfiler: NewDotnetProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.Raw,
							Language:         api.DotNet,
							Tool:             api.DotnetTrace,
							Compressor:       compressor.None,
							Interval:         30 * time.Second,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.DotnetProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				assert.True(t, fields.DotnetProfiler.DotnetManager.(*dotnetManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
		},
		{
			name: "should invoke dotnet-gcdump",
			given: func() (fields, args) {
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						DotnetProfiler: NewDotnetProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.Gcdump,
							Language:         api.DotNet,
							Tool:             api.DotnetGcdump,
							Compressor:       compressor.None,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.DotnetProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				assert.True(t, fields.DotnetProfiler.DotnetManager.(*dotnetManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
			},
		},
		{
			name: "should invoke fail when command fail",
			given: func() (fields, args) {
				commander := executil.NewMockCommander()
				commander.On("Command").Return(&exec.Cmd{})
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						DotnetProfiler: NewDotnetProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.SpeedScope,
							Language:         api.DotNet,
							Tool:             api.DotnetTrace,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.DotnetProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.True(t, fields.DotnetProfiler.DotnetManager.(*dotnetManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
			},
		},
		{
			name: "should invoke fail when fail publish",
			given: func() (fields, args) {
				commander := executil.NewMockCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(errors.New("fake publisher with error"))

				return fields{
						DotnetProfiler: NewDotnetProfiler(commander, publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.SpeedScope,
							Language:         api.DotNet,
							Tool:             api.DotnetTrace,
							Compressor:       compressor.None,
							Interval:         30 * time.Second,
							Iteration:        1,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.DotnetProfiler.invoke(args.job, args.pid)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "fake publisher with error")
				assert.True(t, fields.DotnetProfiler.DotnetManager.(*dotnetManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
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
			tt.then(t, fields, err)

			if tt.after != nil {
				tt.after()
			}
		})
	}
}
