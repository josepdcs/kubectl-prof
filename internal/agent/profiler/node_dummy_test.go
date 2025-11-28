package profiler

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNodeDummyProfiler_SetUp(t *testing.T) {
	type fields struct {
		NodeDummyProfiler *NodeDummyProfiler
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
						NodeDummyProfiler: &NodeDummyProfiler{
							NodeDummyManager: newMockNodeDummyManager(),
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
				return fields.NodeDummyProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				assert.Equal(t, "/root/fs/ContainerID/cwd", fields.NodeDummyProfiler.cwd)
			},
		},
		{
			name: "should fail when get root file system fail",
			given: func() (fields, args) {
				return fields{
						NodeDummyProfiler: &NodeDummyProfiler{
							NodeDummyManager: newMockNodeDummyManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainerWithRootFileSystemLocationResultError,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.NodeDummyProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.NodeDummyProfiler.cwd)
			},
		},
		{
			name: "should fail when get CWD fail",
			given: func() (fields, args) {
				return fields{
						NodeDummyProfiler: &NodeDummyProfiler{
							NodeDummyManager: newMockNodeDummyManager(),
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainerWithCWDResultError,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.NodeDummyProfiler.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.NotNil(t, err)
				assert.Empty(t, fields.NodeDummyProfiler.cwd)
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

func TestNodeDummyProfiler_Invoke(t *testing.T) {
	type fields struct {
		NodeDummyProfiler *NodeDummyProfiler
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
				nodeDummyManager := newMockNodeDummyManager()
				nodeDummyManager.On("invoke", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(nil, time.Duration(0)).Once()

				return fields{
						NodeDummyProfiler: &NodeDummyProfiler{
							NodeDummyManager: nodeDummyManager,
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         time.Duration(0),
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.NodeDummyProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				assert.Nil(t, err)
				fields.NodeDummyProfiler.NodeDummyManager.(*mockNodeDummyManager).AssertNumberOfCalls(t, "invoke", 1)
			},
		},
		{
			name: "should invoke fail when get root PID fail",
			given: func() (fields, args) {
				nodeDummyManager := newMockNodeDummyManager()

				return fields{
						NodeDummyProfiler: &NodeDummyProfiler{
							NodeDummyManager: nodeDummyManager,
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							OutputType:       api.FlameGraph,
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.NodeDummyProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
				assert.EqualError(t, err, "container runtime and container ID are mandatory")
				fields.NodeDummyProfiler.NodeDummyManager.(*mockNodeDummyManager).AssertNumberOfCalls(t, "invoke", 0)
			},
		},
		{
			name: "should invoke fail when invoke fail",
			given: func() (fields, args) {
				nodeDummyManager := newMockNodeDummyManager()
				nodeDummyManager.On("invoke", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(errors.New("fake invoke error"), time.Duration(0)).Once()

				return fields{
						NodeDummyProfiler: &NodeDummyProfiler{
							NodeDummyManager: nodeDummyManager,
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
				return fields.NodeDummyProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				require.Error(t, err)
				assert.EqualError(t, err, "fake invoke error")
				fields.NodeDummyProfiler.NodeDummyManager.(*mockNodeDummyManager).AssertNumberOfCalls(t, "invoke", 1)
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

func TestNodeDummyProfiler_CleanUp(t *testing.T) {
	type fields struct {
		NodeDummyProfiler *NodeDummyProfiler
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
                        NodeDummyProfiler: &NodeDummyProfiler{
                            NodeDummyManager: newMockNodeDummyManager(),
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
				return fields.NodeDummyProfiler.CleanUp(args.job)
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

func Test_nodeDummyManager_invoke(t *testing.T) {
	type fields struct {
		NodeDummyProfiler *NodeDummyProfiler
	}
	type args struct {
		job *job.ProfilingJob
		pid string
		cwd string
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
				cwd := filepath.Join(common.TmpDir(), "cwd")
				_ = os.Mkdir(cwd, os.ModePerm)
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(cwd, config.ProfilingPrefix+"Heap-memory.heapsnapshot"), b.String())

				publisher := publish.NewFakePublisher()
				publisher.On("DoWithNativeGzipAndSplit").Return(nil)

				kill = func(pid, sig int) error {
					return nil
				}

				return fields{
						NodeDummyProfiler: NewNodeDummyProfiler(publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.HeapSnapshot,
							Language:         api.FakeLang,
							Tool:             api.NodeDummy,
							Compressor:       compressor.None,
						},
						pid: "1000",
						cwd: cwd,
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.NodeDummyProfiler.invoke(args.job, args.pid, args.cwd)
			},
			then: func(t *testing.T, fields fields, err error) {
				assert.Nil(t, err)
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"heapsnapshot-1000-0.heapsnapshot")))
				assert.True(t, fields.NodeDummyProfiler.NodeDummyManager.(*nodeDummyManager).publisher.(*publish.Fake).On("DoWithNativeGzipAndSplit").InvokedTimes() == 1)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"heapsnapshot-1000-0.heapsnapshot"))
				_ = file.Remove(filepath.Join(common.TmpDir(), "cwd", config.ProfilingPrefix+"Heap-memory.heapsnapshot"))
				_ = os.Remove(filepath.Join(common.TmpDir(), "cwd"))
			},
		},
		{
			name: "should invoke fail when kill fail",
			given: func() (fields, args) {
				kill = func(pid, sig int) error {
					return errors.New("fake kill error")
				}

				publisher := publish.NewFakePublisher()
				publisher.On("DoWithNativeGzipAndSplit").Return(nil)

				return fields{
						NodeDummyProfiler: NewNodeDummyProfiler(publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.HeapSnapshot,
							Language:         api.FakeLang,
							Tool:             api.NodeDummy,
							Compressor:       compressor.None,
						},
						pid: "1000",
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.NodeDummyProfiler.invoke(args.job, args.pid, args.cwd)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.True(t, fields.NodeDummyProfiler.NodeDummyManager.(*nodeDummyManager).publisher.(*publish.Fake).On("DoWithNativeGzipAndSplit").InvokedTimes() == 0)
			},
		},
		{
			name: "should invoke return fail when file not found",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				cwd := filepath.Join(common.TmpDir(), "cwd")
				_ = os.Mkdir(cwd, os.ModePerm)

				publisher := publish.NewFakePublisher()
				publisher.On("DoWithNativeGzipAndSplit").Return(nil)

				kill = func(pid, sig int) error {
					return nil
				}

				return fields{
						NodeDummyProfiler: &NodeDummyProfiler{
							NodeDummyManager: &nodeDummyManager{
								publisher:       publisher,
								snapshotRetries: 1,
							},
						},
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.FlameGraph,
							Language:         api.Go,
							Tool:             api.NodeDummy,
						},
						pid: "1000",
						cwd: cwd,
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.NodeDummyProfiler.invoke(args.job, args.pid, args.cwd)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.True(t, fields.NodeDummyProfiler.NodeDummyManager.(*nodeDummyManager).publisher.(*publish.Fake).On("DoWithNativeGzipAndSplit").InvokedTimes() == 0)
			},
		},
		{
			name: "should invoke fail when publish result fail",
			given: func() (fields, args) {
				log.SetPrintLogs(true)
				cwd := filepath.Join(common.TmpDir(), "cwd")
				_ = os.Mkdir(cwd, os.ModePerm)
				var b bytes.Buffer
				b.Write([]byte("test"))
				file.Write(filepath.Join(cwd, config.ProfilingPrefix+"Heap-memory.heapsnapshot"), b.String())

				publisher := publish.NewFakePublisher()
				publisher.On("DoWithNativeGzipAndSplit").Return(errors.New("fake publisher with error"))

				kill = func(pid, sig int) error {
					return nil
				}

				return fields{
						NodeDummyProfiler: NewNodeDummyProfiler(publisher),
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
							OutputType:       api.HeapSnapshot,
							Language:         api.FakeLang,
							Tool:             api.NodeDummy,
							Compressor:       compressor.None,
						},
						pid: "1000",
						cwd: cwd,
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.NodeDummyProfiler.invoke(args.job, args.pid, args.cwd)
			},
			then: func(t *testing.T, fields fields, err error) {
				require.Error(t, err)
				assert.ErrorContains(t, err, "fake publisher with error")
				assert.True(t, file.Exists(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"heapsnapshot-1000-0.heapsnapshot")))
				assert.True(t, fields.NodeDummyProfiler.NodeDummyManager.(*nodeDummyManager).publisher.(*publish.Fake).On("DoWithNativeGzipAndSplit").InvokedTimes() == 1)
			},
			after: func() {
				_ = file.Remove(filepath.Join(common.TmpDir(), config.ProfilingPrefix+"heapsnapshot-1000-0.heapsnapshot"))
				_ = file.Remove(filepath.Join(common.TmpDir(), "cwd", config.ProfilingPrefix+"Heap-memory.heapsnapshot"))
				_ = os.Remove(filepath.Join(common.TmpDir(), "cwd"))
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
