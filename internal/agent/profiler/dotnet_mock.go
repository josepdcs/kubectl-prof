package profiler

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/stretchr/testify/mock"
)

// mockDotnetManager is a testify-based mock that implements DotnetManager
type mockDotnetManager struct {
	mock.Mock
}

func newMockDotnetManager() *mockDotnetManager {
	return &mockDotnetManager{}
}

func (m *mockDotnetManager) invoke(j *job.ProfilingJob, pid string) (error, time.Duration) {
	args := m.Called(j, pid)
	var err error
	var d time.Duration
	if a := args.Get(0); a != nil {
		err, _ = a.(error)
	}
	if a := args.Get(1); a != nil {
		d, _ = a.(time.Duration)
	}
	return err, d
}
