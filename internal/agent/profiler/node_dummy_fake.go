package profiler

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
)

// FakeNodeDummyManager is an interface that wraps the NodeDummyManager interface
type FakeNodeDummyManager interface {
	// On sets the expected method to be invoked by the fake NodeDummyManager
	On(methodName string) *fakeNodeDummyManagerMethod

	NodeDummyManager
}

// fakeNodeDummyManager is an implementation of the FakeNodeDummyManager interface
type fakeNodeDummyManager struct {
	fakeMethods map[string]*fakeNodeDummyManagerMethod
}

// fakeMethod represents a fake method
type fakeNodeDummyManagerMethod struct {
	// values to be returned by the fake method
	fakeReturnValues []interface{}
	// number of times the method was invoked
	invokes int
	// index of the execution of the method
	indexExecution int
}

// newFakeNodeDummyManager returns a new fake nodeDummy manager
func newFakeNodeDummyManager() FakeNodeDummyManager {
	return &fakeNodeDummyManager{
		fakeMethods: make(map[string]*fakeNodeDummyManagerMethod),
	}
}

// On sets the expected method to be invoked by the fake publisher
func (f *fakeNodeDummyManager) On(methodName string) *fakeNodeDummyManagerMethod {
	if _, ok := f.fakeMethods[methodName]; !ok {
		f.fakeMethods[methodName] = &fakeNodeDummyManagerMethod{}
	}
	if f.fakeMethods[methodName].fakeReturnValues == nil {
		f.fakeMethods[methodName].fakeReturnValues = make([]interface{}, 0)
	}
	return f.fakeMethods[methodName]
}

// Return sets the values to be returned by the fake method
func (f *fakeNodeDummyManagerMethod) Return(fakeReturnValues ...interface{}) *fakeNodeDummyManagerMethod {
	f.fakeReturnValues = append(f.fakeReturnValues, fakeReturnValues)
	return f
}

// InvokedTimes represents the number of times the method was invoked
func (f *fakeNodeDummyManagerMethod) InvokedTimes() int {
	return f.invokes
}

func (f *fakeNodeDummyManager) invoke(*job.ProfilingJob, string, string) (error, time.Duration) {
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
