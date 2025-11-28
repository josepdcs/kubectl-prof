package profiler

import (
    "bytes"
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
    "github.com/josepdcs/kubectl-prof/pkg/util/log"
    "github.com/pkg/errors"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

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
                            RubyManager: newMockRubyManager(),
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
                            RubyManager: newMockRubyManager(),
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
                            RubyManager: newMockRubyManager(),
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
                rubyManager := newMockRubyManager()
                rubyManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
                    Return(nil, time.Duration(0)).
                    Twice()

                return fields{
                        RubyProfiler: &RubyProfiler{
                            RubyManager: rubyManager,
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
                assert.Nil(t, err)
                fields.RubyProfiler.RubyManager.(*mockRubyManager).AssertNumberOfCalls(t, "invoke", 2)
            },
        },
        {
            name: "should invoke fail when invoke fail",
            given: func() (fields, args) {
                rubyManager := newMockRubyManager()
                rubyManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
                    Return(errors.New("fake invoke error"), time.Duration(0)).
                    Once()

				return fields{

						RubyProfiler: &RubyProfiler{
							RubyManager: rubyManager,
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
                require.Error(t, err)
                assert.EqualError(t, err, "fake invoke error")
                fields.RubyProfiler.RubyManager.(*mockRubyManager).AssertNumberOfCalls(t, "invoke", 1)
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
                            RubyManager: newMockRubyManager(),
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
		then  func(t *testing.T, fields fields, err error)
		after func()
	}{
		{
			name: "should invoke",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"), b.String())

				commander := executil.NewFakeCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						RubyProfiler: NewRubyProfiler(commander, publisher),
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
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg")))
				assert.True(t, fields.RubyProfiler.RubyManager.(*rubyManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
				assert.True(t, fields.RubyProfiler.RubyManager.(*rubyManager).commander.(*executil.Fake).On("Command").InvokedTimes() == 1)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"))
			},
		},
		{
			name: "should invoke fail when command fail",
			given: func() (fields, args) {
				commander := executil.NewFakeCommander()
				commander.On("Command").Return(&exec.Cmd{})
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(nil)

				return fields{
						RubyProfiler: NewRubyProfiler(commander, publisher),
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
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.True(t, fields.RubyProfiler.RubyManager.(*rubyManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 0)
				assert.True(t, fields.RubyProfiler.RubyManager.(*rubyManager).commander.(*executil.Fake).On("Command").InvokedTimes() == 1)
			},
		},
		{
			name: "should fail when publisher fail",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"), b.String())

				commander := executil.NewFakeCommander()
				commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
				publisher := publish.NewFakePublisher()
				publisher.On("Do").Return(errors.New("fake publisher with error"))

				return fields{
						RubyProfiler: NewRubyProfiler(commander, publisher),
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
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "fake publisher with error")
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg")))
				assert.True(t, fields.RubyProfiler.RubyManager.(*rubyManager).publisher.(*publish.Fake).On("Do").InvokedTimes() == 1)
				assert.True(t, fields.RubyProfiler.RubyManager.(*rubyManager).commander.(*executil.Fake).On("Command").InvokedTimes() == 1)

			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"flamegraph-1000.svg"))
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
