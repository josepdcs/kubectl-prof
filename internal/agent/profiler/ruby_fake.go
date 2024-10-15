package profiler

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
)

// FakeRubyManager is an interface that wraps the RubyManager interface
type FakeRubyManager interface {
	// On sets the expected method to be invoked by the fake RubyManager
	On(methodName string) *fakeRubyManagerMethod

	RubyManager
}

// fakeRubyManager is an implementation of the FakeRubyManager interface
type fakeRubyManager struct {
	fakeMethods map[string]*fakeRubyManagerMethod
}

// fakeMethod represents a fake method
type fakeRubyManagerMethod struct {
	// values to be returned by the fake method
	fakeReturnValues []interface{}
	// number of times the method was invoked
	invokes int
	// index of the execution of the method
	indexExecution int
}

// newFakeRubyManager returns a new fake ruby manager
func newFakeRubyManager() FakeRubyManager {
	return &fakeRubyManager{
		fakeMethods: make(map[string]*fakeRubyManagerMethod),
	}
}

// On sets the expected method to be invoked by the fake publisher
func (f *fakeRubyManager) On(methodName string) *fakeRubyManagerMethod {
	if _, ok := f.fakeMethods[methodName]; !ok {
		f.fakeMethods[methodName] = &fakeRubyManagerMethod{}
	}
	if f.fakeMethods[methodName].fakeReturnValues == nil {
		f.fakeMethods[methodName].fakeReturnValues = make([]interface{}, 0)
	}
	return f.fakeMethods[methodName]
}

// Return sets the values to be returned by the fake method
func (f *fakeRubyManagerMethod) Return(fakeReturnValues ...interface{}) *fakeRubyManagerMethod {
	f.fakeReturnValues = append(f.fakeReturnValues, fakeReturnValues)
	return f
}

// InvokedTimes represents the number of times the method was invoked
func (f *fakeRubyManagerMethod) InvokedTimes() int {
	return f.invokes
}

func (f *fakeRubyManager) invoke(*job.ProfilingJob, string) (error, time.Duration) {
	var err error
	var duration time.Duration
	f.fakeMethods["invoke"].invokes++
	if f.fakeMethods["invoke"].fakeReturnValues != nil && len(f.fakeMethods["invoke"].fakeReturnValues) > 0 {
		f.fakeMethods["invoke"].indexExecution++
		arg0 := f.fakeMethods["invoke"].fakeReturnValues[f.fakeMethods["invoke"].indexExecution-1].([]interface{})[0]
		arg1 := f.fakeMethods["invoke"].fakeReturnValues[f.fakeMethods["invoke"].indexExecution-1].([]interface{})[1]
		if arg0 != nil {
			err = arg0.(error)
		}
		if arg1 != nil {
			duration = arg1.(time.Duration)
		}
	}
	return err, duration
}

func (f *fakeRubyManager) handleFlamegraph(*job.ProfilingJob, flamegraph.FrameGrapher, string, string) error {
	var err error
	f.fakeMethods["handleFlamegraph"].invokes++
	if f.fakeMethods["handleFlamegraph"].fakeReturnValues != nil && len(f.fakeMethods["handleFlamegraph"].fakeReturnValues) > 0 {
		f.fakeMethods["handleFlamegraph"].indexExecution++
		arg0 := f.fakeMethods["handleFlamegraph"].fakeReturnValues[f.fakeMethods["handleFlamegraph"].indexExecution-1].([]interface{})[0]
		if arg0 != nil {
			err = arg0.(error)
		}
	}
	return err
}
