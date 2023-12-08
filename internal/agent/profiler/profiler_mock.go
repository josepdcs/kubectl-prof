package profiler

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/pkg/errors"
	"time"
)

type MockProfiler interface {
	Profiler
	SetUpInvokedTimes() int
	InvokeInvokedTimes() int
	CleanUpInvokedTimes() int
}

type DefaultMockProfiler struct {
	setUpInvokedTimes   int
	invokeInvokedTimes  int
	cleanUpInvokedTimes int
}

func NewMockProfiler() *DefaultMockProfiler {
	return &DefaultMockProfiler{}
}

func (m *DefaultMockProfiler) SetUp(job *job.ProfilingJob) error {
	m.setUpInvokedTimes++
	if job.ContainerID == "WithSetupError" {
		return errors.New("fake SetUp with error")
	}
	fmt.Println("fake SetUp")
	return nil
}

func (m *DefaultMockProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()
	m.invokeInvokedTimes++
	if job.ContainerID == "WithInvokeError" {
		return errors.New("fake Invoke with error"), time.Since(start)
	}
	fmt.Println("fake Invoke")
	return nil, time.Since(start)
}

func (m *DefaultMockProfiler) CleanUp(*job.ProfilingJob) error {
	m.cleanUpInvokedTimes++
	fmt.Println("fake CleanUp")
	return nil
}

func (m *DefaultMockProfiler) SetUpInvokedTimes() int {
	return m.setUpInvokedTimes
}

func (m *DefaultMockProfiler) InvokeInvokedTimes() int {
	return m.invokeInvokedTimes
}

func (m *DefaultMockProfiler) CleanUpInvokedTimes() int {
	return m.cleanUpInvokedTimes
}
