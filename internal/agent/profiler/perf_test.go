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
)

type mockPerfUtil struct {
	mock.Mock
}

func (m *mockPerfUtil) runPerfRecord(job *job.ProfilingJob) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *mockPerfUtil) runPerfScript(job *job.ProfilingJob) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *mockPerfUtil) foldPerfOutput(job *job.ProfilingJob) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *mockPerfUtil) generateFlameGraph(job *job.ProfilingJob) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *mockPerfUtil) publishResult(c compressor.Type, fileName string, outputType api.OutputType) error {
	args := m.Called(c, fileName, outputType)
	return args.Error(0)
}

func TestNewPerfProfiler(t *testing.T) {
	p := NewPerfProfiler()
	assert.IsType(t, p, &PerfProfiler{})
}

func TestPerfProfiler_SetUp(t *testing.T) {
	p := NewPerfProfiler()
	j := &job.ProfilingJob{}
	assert.Equal(t, nil, p.SetUp(j))
}

func TestPerfProfiler_Invoke(t *testing.T) {
	type args struct {
		job *job.ProfilingJob
		p   Profiler
	}
	tests := []struct {
		name       string
		args       args
		mock       func(m *mockPerfUtil)
		assert     func(m *mockPerfUtil)
		wantErrMsg string
	}{
		{
			name: "Invoke should publish result",
			args: args{
				job: &job.ProfilingJob{},
			},
			mock: func(m *mockPerfUtil) {
				m.On("runPerfRecord", mock.Anything).Return(nil)
				m.On("runPerfScript", mock.Anything).Return(nil)
				m.On("foldPerfOutput", mock.Anything).Return(nil)
				m.On("generateFlameGraph", mock.Anything).Return(nil)
				m.On("publishResult", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			assert: func(m *mockPerfUtil) {
				m.AssertNumberOfCalls(t, "runPerfRecord", 1)
				m.AssertNumberOfCalls(t, "runPerfScript", 1)
				m.AssertNumberOfCalls(t, "foldPerfOutput", 1)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 1)
				m.AssertNumberOfCalls(t, "publishResult", 1)
			},
		},
		{
			name: "Invoke should fail when run perf record fail",
			args: args{
				job: &job.ProfilingJob{},
			},
			mock: func(m *mockPerfUtil) {
				m.On("runPerfRecord", mock.Anything).Return(errors.New("error"))
			},
			assert: func(m *mockPerfUtil) {
				m.AssertNumberOfCalls(t, "runPerfRecord", 1)
				m.AssertNumberOfCalls(t, "runPerfScript", 0)
				m.AssertNumberOfCalls(t, "foldPerfOutput", 0)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 0)
				m.AssertNumberOfCalls(t, "publishResult", 0)
			},
			wantErrMsg: "perf record failed: error",
		},
		{
			name: "Invoke should fail when run perf script fail",
			args: args{
				job: &job.ProfilingJob{},
			},
			mock: func(m *mockPerfUtil) {
				m.On("runPerfRecord", mock.Anything).Return(nil)
				m.On("runPerfScript", mock.Anything).Return(errors.New("error"))
			},
			assert: func(m *mockPerfUtil) {
				m.AssertNumberOfCalls(t, "runPerfRecord", 1)
				m.AssertNumberOfCalls(t, "runPerfScript", 1)
				m.AssertNumberOfCalls(t, "foldPerfOutput", 0)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 0)
				m.AssertNumberOfCalls(t, "publishResult", 0)
			},
			wantErrMsg: "perf script failed: error",
		},
		{
			name: "Invoke should fail when folder perf output fail",
			args: args{
				job: &job.ProfilingJob{},
			},
			mock: func(m *mockPerfUtil) {
				m.On("runPerfRecord", mock.Anything).Return(nil)
				m.On("runPerfScript", mock.Anything).Return(nil)
				m.On("foldPerfOutput", mock.Anything).Return(errors.New("error"))
			},
			assert: func(m *mockPerfUtil) {
				m.AssertNumberOfCalls(t, "runPerfRecord", 1)
				m.AssertNumberOfCalls(t, "runPerfScript", 1)
				m.AssertNumberOfCalls(t, "foldPerfOutput", 1)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 0)
				m.AssertNumberOfCalls(t, "publishResult", 0)
			},
			wantErrMsg: "folding perf output failed: error",
		},
		{
			name: "Invoke should fail when generate FlameGraph fail",
			args: args{
				job: &job.ProfilingJob{},
			},
			mock: func(m *mockPerfUtil) {
				m.On("runProfiler", mock.Anything).Return(nil)
				m.On("runPerfRecord", mock.Anything).Return(nil)
				m.On("runPerfScript", mock.Anything).Return(nil)
				m.On("foldPerfOutput", mock.Anything).Return(nil)
				m.On("generateFlameGraph", mock.Anything).Return(errors.New("error"))
			},
			assert: func(m *mockPerfUtil) {
				m.AssertNumberOfCalls(t, "runPerfRecord", 1)
				m.AssertNumberOfCalls(t, "runPerfScript", 1)
				m.AssertNumberOfCalls(t, "foldPerfOutput", 1)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 1)
				m.AssertNumberOfCalls(t, "publishResult", 0)
			},
			wantErrMsg: "flamegraph generation failed: error",
		},
		{
			name: "Invoke should fail when publish result fail",
			args: args{
				job: &job.ProfilingJob{},
			},
			mock: func(m *mockPerfUtil) {
				m.On("runProfiler", mock.Anything).Return(nil)
				m.On("runPerfRecord", mock.Anything).Return(nil)
				m.On("runPerfScript", mock.Anything).Return(nil)
				m.On("foldPerfOutput", mock.Anything).Return(nil)
				m.On("generateFlameGraph", mock.Anything).Return(nil)
				m.On("publishResult", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			assert: func(m *mockPerfUtil) {
				m.AssertNumberOfCalls(t, "runPerfRecord", 1)
				m.AssertNumberOfCalls(t, "runPerfScript", 1)
				m.AssertNumberOfCalls(t, "foldPerfOutput", 1)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 1)
				m.AssertNumberOfCalls(t, "publishResult", 1)
			},
			wantErrMsg: "error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockPerfUtil{}
			b := &PerfProfiler{
				PerfUtil(m),
			}
			tt.mock(m)

			err, _ := b.Invoke(tt.args.job)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				require.NoError(t, err)
			}

			tt.assert(m)
		})
	}
}
