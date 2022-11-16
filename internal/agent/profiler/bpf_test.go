package profiler

import (
	"errors"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type mockBpfManager struct {
	mock.Mock
}

func (m *mockBpfManager) runProfiler(job *job.ProfilingJob, targetPID string) error {
	args := m.Called(job, targetPID)
	return args.Error(0)
}

func (m *mockBpfManager) generateFlameGraph(fileName string) error {
	args := m.Called(fileName)
	return args.Error(0)
}

func (m *mockBpfManager) moveSources(target string) error {
	args := m.Called(target)
	return args.Error(0)
}

func (m *mockBpfManager) publishResult(c compressor.Type, fileName string, outputType api.EventType) error {
	args := m.Called(c, fileName, outputType)
	return args.Error(0)
}

func (m *mockBpfManager) cleanUp() {

}

func TestNewBpfProfiler(t *testing.T) {
	p := NewBpfProfiler()
	assert.IsType(t, p, &BpfProfiler{})
}

func TestBpfProfiler_SetUp(t *testing.T) {
	type fields struct {
		*BpfProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		args  args
		given func() (fields, args)
		when  func(fields, args) error
		then  func(t *testing.T, err error, fields fields)
	}{
		{
			name: "should setup",
			args: args{
				job: &job.ProfilingJob{},
			},
			given: func() (fields, args) {
				m := &mockBpfManager{}
				b := &BpfProfiler{
					BpfManager: BpfManager(m),
				}
				return fields{
						b,
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainer,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				//mock := fields.BpfProfiler.BpfManager.(*mockBpfManager)
				require.NoError(t, err)
				assert.Equal(t, "PID_ContainerID", fields.targetPID)
			},
		},
		{
			name: "should fail when container PID not found",
			args: args{
				job: &job.ProfilingJob{},
			},
			given: func() (fields, args) {
				m := &mockBpfManager{}
				b := &BpfProfiler{
					BpfManager: BpfManager(m),
				}
				return fields{
						b,
					}, args{
						job: &job.ProfilingJob{
							Duration:         0,
							ContainerRuntime: api.FakeContainerWithPIDResultError,
							ContainerID:      "ContainerID",
						},
					}
			},
			when: func(fields fields, args args) error {
				return fields.SetUp(args.job)
			},
			then: func(t *testing.T, err error, fields fields) {
				//mock := fields.BpfProfiler.BpfManager.(*mockBpfManager)
				require.Error(t, err)
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
		BpfProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		args  args
		given func() (fields, args)
		when  func(fields, args) (error, time.Duration)
		then  func(t *testing.T, err error, mock *mockBpfManager)
	}{
		{
			name: "should publish result",
			args: args{
				job: &job.ProfilingJob{},
			},
			given: func() (fields, args) {
				m := &mockBpfManager{}
				b := BpfProfiler{
					BpfManager: BpfManager(m),
				}
				m.On("runProfiler", mock.Anything, mock.Anything).Return(nil)
				m.On("generateFlameGraph", mock.Anything).Return(nil)
				m.On("publishResult", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return fields{
						b,
					}, args{
						job: &job.ProfilingJob{},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.Invoke(args.job)
			},
			then: func(t *testing.T, err error, mock *mockBpfManager) {
				require.NoError(t, err)
				mock.AssertNumberOfCalls(t, "runProfiler", 1)
				mock.AssertNumberOfCalls(t, "generateFlameGraph", 1)
				mock.AssertNumberOfCalls(t, "publishResult", 1)
			},
		},
		{
			name: "should fail when run profiler fail",
			args: args{
				job: &job.ProfilingJob{},
			},
			given: func() (fields, args) {
				m := &mockBpfManager{}
				b := BpfProfiler{
					BpfManager: BpfManager(m),
				}
				m.On("runProfiler", mock.Anything, mock.Anything).Return(errors.New("error"))
				return fields{
						b,
					}, args{
						job: &job.ProfilingJob{},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.Invoke(args.job)
			},
			then: func(t *testing.T, err error, mock *mockBpfManager) {
				require.Error(t, err)
				mock.AssertNumberOfCalls(t, "runProfiler", 1)
				mock.AssertNumberOfCalls(t, "generateFlameGraph", 0)
				mock.AssertNumberOfCalls(t, "publishResult", 0)
			},
		},
		{
			name: "should fail when generate FlameGraph fail",
			args: args{
				job: &job.ProfilingJob{},
			},
			given: func() (fields, args) {
				m := &mockBpfManager{}
				b := BpfProfiler{
					BpfManager: BpfManager(m),
				}
				m.On("runProfiler", mock.Anything, mock.Anything).Return(nil)
				m.On("generateFlameGraph", mock.Anything).Return(errors.New("error"))
				return fields{
						b,
					}, args{
						job: &job.ProfilingJob{},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.Invoke(args.job)
			},
			then: func(t *testing.T, err error, mock *mockBpfManager) {
				require.Error(t, err)
				mock.AssertNumberOfCalls(t, "runProfiler", 1)
				mock.AssertNumberOfCalls(t, "generateFlameGraph", 1)
				mock.AssertNumberOfCalls(t, "publishResult", 0)
			},
		},
		{
			name: "should fail when publish result fail",
			args: args{
				job: &job.ProfilingJob{},
			},
			given: func() (fields, args) {
				m := &mockBpfManager{}
				b := BpfProfiler{
					BpfManager: BpfManager(m),
				}
				m.On("runProfiler", mock.Anything, mock.Anything).Return(nil)
				m.On("generateFlameGraph", mock.Anything).Return(nil)
				m.On("publishResult", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
				return fields{
						b,
					}, args{
						job: &job.ProfilingJob{},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.Invoke(args.job)
			},
			then: func(t *testing.T, err error, mock *mockBpfManager) {
				require.Error(t, err)
				mock.AssertNumberOfCalls(t, "runProfiler", 1)
				mock.AssertNumberOfCalls(t, "generateFlameGraph", 1)
				mock.AssertNumberOfCalls(t, "publishResult", 1)
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
			tt.then(t, err, fields.BpfProfiler.BpfManager.(*mockBpfManager))
		})
	}
}
