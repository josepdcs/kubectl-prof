package profiler

import (
    "time"

    "github.com/josepdcs/kubectl-prof/internal/agent/job"
    "github.com/stretchr/testify/mock"
)

// mockRubyManager is a testify-based mock that implements RubyManager
type mockRubyManager struct {
    mock.Mock
}

func newMockRubyManager() *mockRubyManager {
    return &mockRubyManager{}
}

func (m *mockRubyManager) invoke(j *job.ProfilingJob, pid string) (error, time.Duration) {
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
