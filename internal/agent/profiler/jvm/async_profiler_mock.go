package jvm

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/stretchr/testify/mock"
)

// mockAsyncProfilerManager is a testify-based mock that implements AsyncProfilerManager
type mockAsyncProfilerManager struct {
	mock.Mock
}

func newMockAsyncProfilerManager() *mockAsyncProfilerManager {
	return &mockAsyncProfilerManager{}
}

func (m *mockAsyncProfilerManager) stageProfilerLibrary(targetFs string) error {
	args := m.Called(targetFs)
	if a := args.Get(0); a != nil {
		if err, ok := a.(error); ok {
			return err
		}
	}
	return nil
}

func (m *mockAsyncProfilerManager) releaseTargetArtifacts() {
	_ = m.Called()
}

func (m *mockAsyncProfilerManager) invoke(j *job.ProfilingJob, pid string) (error, time.Duration) {
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

func (m *mockAsyncProfilerManager) cleanUp(j *job.ProfilingJob, pid string) {
	_ = m.Called(j, pid)
}
