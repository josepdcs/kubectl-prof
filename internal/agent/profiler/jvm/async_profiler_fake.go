package jvm

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
)

// FakeAsyncProfilerManager is an interface that wraps the AsyncProfilerManager interface
type FakeAsyncProfilerManager interface {
	// On sets the expected method to be invoked by the fake AsyncProfilerManager
	On(methodName string) *fakeAsyncProfilerManagerMethod

	AsyncProfilerManager
}

// fakeAsyncProfilerManager is an implementation of the FakeAsyncProfilerManager interface
type fakeAsyncProfilerManager struct {
	fakeMethods map[string]*fakeAsyncProfilerManagerMethod
}

// fakeMethod represents a fake method
type fakeAsyncProfilerManagerMethod struct {
	// values to be returned by the fake method
	fakeReturnValues []interface{}
	// number of times the method was invoked
	invokes int
	// index of the execution of the method
	indexExecution int
}

// newFakeAsyncProfilerManager returns a new fake asyncProfiler manager
func newFakeAsyncProfilerManager() FakeAsyncProfilerManager {
	return &fakeAsyncProfilerManager{
		fakeMethods: make(map[string]*fakeAsyncProfilerManagerMethod),
	}
}

// On sets the expected method to be invoked by the fake publisher
func (f *fakeAsyncProfilerManager) On(methodName string) *fakeAsyncProfilerManagerMethod {
	if _, ok := f.fakeMethods[methodName]; !ok {
		f.fakeMethods[methodName] = &fakeAsyncProfilerManagerMethod{}
	}
	if f.fakeMethods[methodName].fakeReturnValues == nil {
		f.fakeMethods[methodName].fakeReturnValues = make([]interface{}, 0)
	}
	return f.fakeMethods[methodName]
}

// Return sets the values to be returned by the fake method
func (f *fakeAsyncProfilerManagerMethod) Return(fakeReturnValues ...interface{}) *fakeAsyncProfilerManagerMethod {
	f.fakeReturnValues = append(f.fakeReturnValues, fakeReturnValues)
	return f
}

// InvokedTimes represents the number of times the method was invoked
func (f *fakeAsyncProfilerManagerMethod) InvokedTimes() int {
	return f.invokes
}

func (f *fakeAsyncProfilerManager) invoke(*job.ProfilingJob, string) (error, time.Duration) {
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

func (f *fakeAsyncProfilerManager) getTmpDir() string {
	return "/tmp"
}

func (f *fakeAsyncProfilerManager) removeTmpDir() error {
	var err error
	f.fakeMethods["removeTmpDir"].invokes++
	if f.fakeMethods["removeTmpDir"].fakeReturnValues != nil && len(f.fakeMethods["removeTmpDir"].fakeReturnValues) > 0 {
		f.fakeMethods["removeTmpDir"].indexExecution++
		arg0 := f.fakeMethods["removeTmpDir"].fakeReturnValues[f.fakeMethods["removeTmpDir"].indexExecution-1].([]interface{})[0]
		if arg0 != nil {
			err = arg0.(error)
		}
	}
	return err
}

func (f *fakeAsyncProfilerManager) linkTmpDirToTargetTmpDir(s string) error {
	var err error
	f.fakeMethods["linkTmpDirToTargetTmpDir"].invokes++
	if f.fakeMethods["linkTmpDirToTargetTmpDir"].fakeReturnValues != nil && len(f.fakeMethods["linkTmpDirToTargetTmpDir"].fakeReturnValues) > 0 {
		f.fakeMethods["linkTmpDirToTargetTmpDir"].indexExecution++
		arg0 := f.fakeMethods["linkTmpDirToTargetTmpDir"].fakeReturnValues[f.fakeMethods["linkTmpDirToTargetTmpDir"].indexExecution-1].([]interface{})[0]
		if arg0 != nil {
			err = arg0.(error)
		}
	}
	return err
}

func (f *fakeAsyncProfilerManager) copyProfilerToTmpDir() error {
	var err error
	f.fakeMethods["copyProfilerToTmpDir"].invokes++
	if f.fakeMethods["copyProfilerToTmpDir"].fakeReturnValues != nil && len(f.fakeMethods["copyProfilerToTmpDir"].fakeReturnValues) > 0 {
		f.fakeMethods["copyProfilerToTmpDir"].indexExecution++
		arg0 := f.fakeMethods["copyProfilerToTmpDir"].fakeReturnValues[f.fakeMethods["copyProfilerToTmpDir"].indexExecution-1].([]interface{})[0]
		if arg0 != nil {
			err = arg0.(error)
		}
	}
	return err
}

func (f *fakeAsyncProfilerManager) cleanUp(profilingJob *job.ProfilingJob, s string) {
	f.fakeMethods["cleanUp"].invokes++
	if f.fakeMethods["cleanUp"].fakeReturnValues != nil && len(f.fakeMethods["cleanUp"].fakeReturnValues) > 0 {
		f.fakeMethods["cleanUp"].indexExecution++
	}
}
