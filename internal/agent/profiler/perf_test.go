package profiler

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	executil "github.com/josepdcs/kubectl-prof/internal/agent/util/exec"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	p := NewPerfProfiler(executil.NewCommander(), publish.NewPublisher())
	assert.IsType(t, p, &PerfProfiler{})
}
