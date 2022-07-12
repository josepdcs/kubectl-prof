package profiler

import (
	"errors"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockBpfUtil struct {
	mock.Mock
}

func (m *mockBpfUtil) runProfiler(job *config.ProfilingJob) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *mockBpfUtil) generateFlameGraph(fileName string) error {
	args := m.Called(fileName)
	return args.Error(0)
}

func (m *mockBpfUtil) moveSources(target string) error {
	args := m.Called(target)
	return args.Error(0)
}

func (m *mockBpfUtil) publishResult(c api.Compressor, fileName string, outputType api.EventType) error {
	args := m.Called(c, fileName, outputType)
	return args.Error(0)
}

func TestNewBpfProfiler(t *testing.T) {
	p := NewBpfProfiler()
	assert.IsType(t, p, &BpfProfiler{})
}

func TestBpfProfiler_SetUp(t *testing.T) {
	type args struct {
		job *config.ProfilingJob
		p   Profiler
	}
	tests := []struct {
		name       string
		args       args
		mock       func(m *mockBpfUtil)
		assert     func(m *mockBpfUtil)
		wantErrMsg string
	}{
		{
			name: "SetUp should move resources",
			args: args{
				job: &config.ProfilingJob{},
			},
			mock: func(m *mockBpfUtil) {
				m.On("moveSources", mock.Anything, mock.Anything).Return(nil)
			},
			assert: func(m *mockBpfUtil) {
				m.AssertNumberOfCalls(t, "moveSources", 1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockBpfUtil{}
			b := &BpfProfiler{
				BpfUtil(m),
			}
			tt.mock(m)

			err := b.SetUp(tt.args.job)

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

func TestBpfProfiler_Invoke(t *testing.T) {
	type args struct {
		job *config.ProfilingJob
		p   Profiler
	}
	tests := []struct {
		name       string
		args       args
		mock       func(m *mockBpfUtil)
		assert     func(m *mockBpfUtil)
		wantErrMsg string
	}{
		{
			name: "Invoke should publish result",
			args: args{
				job: &config.ProfilingJob{},
			},
			mock: func(m *mockBpfUtil) {
				m.On("runProfiler", mock.Anything).Return(nil)
				m.On("generateFlameGraph", mock.Anything).Return(nil)
				m.On("publishResult", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			assert: func(m *mockBpfUtil) {
				m.AssertNumberOfCalls(t, "runProfiler", 1)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 1)
				m.AssertNumberOfCalls(t, "publishResult", 1)
			},
		},
		{
			name: "Invoke should fail when run profiler fail",
			args: args{
				job: &config.ProfilingJob{},
			},
			mock: func(m *mockBpfUtil) {
				m.On("runProfiler", mock.Anything).Return(errors.New("error"))
			},
			assert: func(m *mockBpfUtil) {
				m.AssertNumberOfCalls(t, "runProfiler", 1)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 0)
				m.AssertNumberOfCalls(t, "publishResult", 0)
			},
			wantErrMsg: "profiling failed: error",
		},
		{
			name: "Invoke should fail when generate FlameGraph fail",
			args: args{
				job: &config.ProfilingJob{},
			},
			mock: func(m *mockBpfUtil) {
				m.On("runProfiler", mock.Anything).Return(nil)
				m.On("generateFlameGraph", mock.Anything).Return(errors.New("error"))
			},
			assert: func(m *mockBpfUtil) {
				m.AssertNumberOfCalls(t, "runProfiler", 1)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 1)
				m.AssertNumberOfCalls(t, "publishResult", 0)
			},
			wantErrMsg: "flamegraph generation failed: error",
		},
		{
			name: "Invoke should fail when publish result fail",
			args: args{
				job: &config.ProfilingJob{},
			},
			mock: func(m *mockBpfUtil) {
				m.On("runProfiler", mock.Anything).Return(nil)
				m.On("generateFlameGraph", mock.Anything).Return(nil)
				m.On("publishResult", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			assert: func(m *mockBpfUtil) {
				m.AssertNumberOfCalls(t, "runProfiler", 1)
				m.AssertNumberOfCalls(t, "generateFlameGraph", 1)
				m.AssertNumberOfCalls(t, "publishResult", 1)
			},
			wantErrMsg: "error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockBpfUtil{}
			b := &BpfProfiler{
				BpfUtil(m),
			}
			tt.mock(m)

			err := b.Invoke(tt.args.job)

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
