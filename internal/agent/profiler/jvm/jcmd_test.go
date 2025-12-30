package jvm

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
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
				jcmdManager := newMockJcmdManager()
				jcmdManager.On("removeTmpDir").Return(nil)
				jcmdManager.On("linkTmpDirToTargetTmpDir", mock.AnythingOfType("string")).Return(nil)
				jcmdManager.On("copyJfrSettingsToTmpDir").Return(nil)

				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: jcmdManager,
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
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.JcmdProfiler.targetPIDs)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "removeTmpDir", 1)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "linkTmpDirToTargetTmpDir", 1)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "copyJfrSettingsToTmpDir", 1)
			},
		},
		{
			name: "should setup when PID is given",
			given: func() (fields, args) {
				jcmdManager := newMockJcmdManager()
				jcmdManager.On("removeTmpDir").Return(nil)
				jcmdManager.On("linkTmpDirToTargetTmpDir", mock.AnythingOfType("string")).Return(nil)
				jcmdManager.On("copyJfrSettingsToTmpDir").Return(nil)

				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: jcmdManager,
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
				assert.Nil(t, err)
				assert.Equal(t, []string{"PID_ContainerID"}, fields.JcmdProfiler.targetPIDs)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "removeTmpDir", 1)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "linkTmpDirToTargetTmpDir", 1)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "copyJfrSettingsToTmpDir", 1)
			},
		},
		{
			name: "should fail when getting target filesystem fail",
			given: func() (fields, args) {
				jcmdManager := newMockJcmdManager()
				jcmdManager.On("removeTmpDir").Return(nil)
				jcmdManager.On("linkTmpDirToTargetTmpDir", mock.AnythingOfType("string")).Return(nil)
				jcmdManager.On("copyJfrSettingsToTmpDir").Return(nil)

				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: jcmdManager,
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
				assert.NotNil(t, err)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "removeTmpDir", 0)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "linkTmpDirToTargetTmpDir", 0)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "copyJfrSettingsToTmpDir", 0)
			},
		},
		{
			name: "should fail when removing tmp dir fail",
			given: func() (fields, args) {
				jcmdManager := newMockJcmdManager()
				jcmdManager.On("removeTmpDir").Return(errors.New("fake error"))

				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: jcmdManager,
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
				assert.NotNil(t, err)
				assert.EqualError(t, err, "fake error")
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "removeTmpDir", 1)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "linkTmpDirToTargetTmpDir", 0)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "copyJfrSettingsToTmpDir", 0)
			},
		},
		{
			name: "should fail when link tmp dir to target tmp dir fail",
			given: func() (fields, args) {
				jcmdManager := newMockJcmdManager()
				jcmdManager.On("removeTmpDir").Return(nil)
				jcmdManager.On("linkTmpDirToTargetTmpDir", mock.AnythingOfType("string")).Return(errors.New("fake error"))

				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: jcmdManager,
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
				assert.NotNil(t, err)
				assert.EqualError(t, err, "fake error")
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "removeTmpDir", 1)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "linkTmpDirToTargetTmpDir", 1)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "copyJfrSettingsToTmpDir", 0)
			},
		},
		{
			name: "should fail when container PID not found",
			given: func() (fields, args) {
				jcmdManager := newMockJcmdManager()
				jcmdManager.On("removeTmpDir").Return(nil)
				jcmdManager.On("linkTmpDirToTargetTmpDir", mock.AnythingOfType("string")).Return(nil)

				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: jcmdManager,
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
				assert.NotNil(t, err)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "removeTmpDir", 1)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "linkTmpDirToTargetTmpDir", 1)
				fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "copyJfrSettingsToTmpDir", 0)
			},
		},
  {
            name: "should fail when copy jfr settings to tmp dir fail",
            given: func() (fields, args) {
                jcmdManager := newMockJcmdManager()
                jcmdManager.On("removeTmpDir").Return(nil)
                jcmdManager.On("linkTmpDirToTargetTmpDir", mock.AnythingOfType("string")).Return(nil)
                jcmdManager.On("copyJfrSettingsToTmpDir").Return(errors.New("fake error"))

				return fields{
						JcmdProfiler: &JcmdProfiler{
							JcmdManager: jcmdManager,
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
                assert.NotNil(t, err)
                assert.EqualError(t, err, "fake error")
                fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "removeTmpDir", 1)
                fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "linkTmpDirToTargetTmpDir", 1)
                fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "copyJfrSettingsToTmpDir", 1)
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
                jcmdManager := newMockJcmdManager()
                jcmdManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
                    Return(nil, time.Duration(0)).Twice()

				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: jcmdManager,
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
				fields.JcmdProfiler.delay = 0
				fields.JcmdProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.JcmdProfiler.Invoke(args.job)
			},
            then: func(t *testing.T, err error, fields fields) {
                assert.Nil(t, err)
                fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "invoke", 2)
            },
        },
        {
            name: "should invoke fail when invoke fail",
            given: func() (fields, args) {
                jcmdManager := newMockJcmdManager()
                jcmdManager.On("invoke", mock.Anything, mock.AnythingOfType("string")).
                    Return(errors.New("fake invoke error"), time.Duration(0)).Once()

				return fields{
						JcmdProfiler: JcmdProfiler{
							JcmdManager: jcmdManager,
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
				fields.JcmdProfiler.delay = 0
				fields.JcmdProfiler.targetPIDs = []string{"1000", "2000"}
				return fields.JcmdProfiler.Invoke(args.job)
			},
            then: func(t *testing.T, err error, fields fields) {
                require.Error(t, err)
                assert.EqualError(t, err, "fake invoke error")
                fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "invoke", 1)
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
		JcmdProfiler JcmdProfiler
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
				recordingPIDs = make(chan string, 2)
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"jfr.jfr")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])

    jcmdManager := newMockJcmdManager()
    jcmdManager.On("cleanUp", mock.Anything, "1000").Return().Once()
    jcmdManager.On("cleanUp", mock.Anything, "2000").Return().Once()

				return fields{
						JcmdProfiler: JcmdProfiler{
							targetPIDs:  []string{"1000", "2000"},
							delay:       0,
							JcmdManager: jcmdManager,
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:            0,
							ContainerRuntime:    api.FakeContainer,
							ContainerID:         "ContainerID",
							Compressor:          compressor.Gzip,
							OutputType:          api.Jfr,
							AdditionalArguments: map[string]string{"settings": "contprof"},
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.JcmdProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"jfr.jfr")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"jfr.jfr"+
					compressor.GetExtensionFileByCompressor[compressor.Gzip])
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
                assert.Nil(t, err)
                fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "cleanUp", 2)
            },
        },
        {
            name: "should clean up when not jfr output type",
            given: func() (fields, args) {
				recordingPIDs = make(chan string, 2)
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"heapdump.hprof")
				_, _ = os.Create(f)
				_, _ = os.Create(f + compressor.GetExtensionFileByCompressor[compressor.Gzip])

    jcmdManager := newMockJcmdManager()

				return fields{
						JcmdProfiler: JcmdProfiler{
							targetPIDs:  []string{"1000", "2000"},
							delay:       0,
							JcmdManager: jcmdManager,
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:            0,
							ContainerRuntime:    api.FakeContainer,
							ContainerID:         "ContainerID",
							Compressor:          compressor.Gzip,
							OutputType:          api.HeapDump,
							AdditionalArguments: map[string]string{"settings": "contprof"},
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.JcmdProfiler.CleanUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				f := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"heapdump.hprof")
				g := filepath.Join(common.TmpDir(), config.ProfilingPrefix+"heapdump.hprof"+
					compressor.GetExtensionFileByCompressor[compressor.Gzip])
				assert.False(t, file.Exists(f))
				assert.False(t, file.Exists(g))
                assert.Nil(t, err)
                fields.JcmdProfiler.JcmdManager.(*mockJcmdManager).AssertNumberOfCalls(t, "cleanUp", 0)
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

func Test_jcmdManager_publishResult(t *testing.T) {
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
						JcmdProfiler: *NewJcmdProfiler(executil.NewCommander(), publish.NewPublisher()),
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
						JcmdProfiler: *NewJcmdProfiler(executil.NewCommander(), publish.NewPublisher()),
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

func Test_jcmdManager_cleanUp(t *testing.T) {
	recordingPIDs = make(chan string, 2)

	commander := executil.NewFakeCommander()
	commander.On("Command").Return(exec.Command("ls", common.TmpDir()))
	publisher := publish.NewFakePublisher()
	a := NewJcmdProfiler(commander, publisher)
	a.cleanUp(&job.ProfilingJob{
		Duration:         0,
		ContainerRuntime: api.FakeContainer,
		ContainerID:      "ContainerID",
		OutputType:       api.Jfr,
		Language:         api.FakeLang,
		Tool:             api.Jcmd,
		Compressor:       compressor.None,
	}, "1000")
	assert.True(t, commander.(*executil.Fake).On("Command").InvokedTimes() == 1)

	commander = executil.NewFakeCommander()
	commander.On("Command").Return(&exec.Cmd{})
	a = NewJcmdProfiler(commander, publisher)
	a.cleanUp(&job.ProfilingJob{
		Duration:         0,
		ContainerRuntime: api.FakeContainer,
		ContainerID:      "ContainerID",
		OutputType:       api.Jfr,
		Language:         api.FakeLang,
		Tool:             api.Jcmd,
		Compressor:       compressor.None,
	}, "1000")
	assert.True(t, commander.(*executil.Fake).On("Command").InvokedTimes() == 1)
}
