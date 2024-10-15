package profiler

import (
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/flamegraph"
)

// FakePerfManager is an interface that wraps the PerfManager interface
type FakePerfManager interface {
	// On sets the expected method to be invoked by the fake PerfManager
	On(methodName string) *fakePerfManagerMethod

	PerfManager
}

// fakePerfManager is an implementation of the FakePerfManager interface
type fakePerfManager struct {
	fakeMethods map[string]*fakePerfManagerMethod
}

// fakeMethod represents a fake method
type fakePerfManagerMethod struct {
	// values to be returned by the fake method
	fakeReturnValues []interface{}
	// number of times the method was invoked
	invokes int
	// index of the execution of the method
	indexExecution int
}

// newFakePerfManager returns a new fake perf manager
func newFakePerfManager() FakePerfManager {
	return &fakePerfManager{
		fakeMethods: make(map[string]*fakePerfManagerMethod),
	}
}

// On sets the expected method to be invoked by the fake publisher
func (f *fakePerfManager) On(methodName string) *fakePerfManagerMethod {
	if _, ok := f.fakeMethods[methodName]; !ok {
		f.fakeMethods[methodName] = &fakePerfManagerMethod{}
	}
	if f.fakeMethods[methodName].fakeReturnValues == nil {
		f.fakeMethods[methodName].fakeReturnValues = make([]interface{}, 0)
	}
	return f.fakeMethods[methodName]
}

// Return sets the values to be returned by the fake method
func (f *fakePerfManagerMethod) Return(fakeReturnValues ...interface{}) *fakePerfManagerMethod {
	f.fakeReturnValues = append(f.fakeReturnValues, fakeReturnValues)
	return f
}

// InvokedTimes represents the number of times the method was invoked
func (f *fakePerfManagerMethod) InvokedTimes() int {
	return f.invokes
}

func (f *fakePerfManager) invoke(*job.ProfilingJob, string) (error, time.Duration) {
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

func (f *fakePerfManager) runPerfRecord(*job.ProfilingJob, string) error {
	var err error
	f.fakeMethods["runPerfRecord"].invokes++
	if f.fakeMethods["runPerfRecord"].fakeReturnValues != nil && len(f.fakeMethods["runPerfRecord"].fakeReturnValues) > 0 {
		f.fakeMethods["runPerfRecord"].indexExecution++
		arg0 := f.fakeMethods["runPerfRecord"].fakeReturnValues[f.fakeMethods["runPerfRecord"].indexExecution-1].([]interface{})[0]
		if arg0 != nil {
			err = arg0.(error)
		}
	}
	return err
}

func (f *fakePerfManager) runPerfScript(*job.ProfilingJob, string) error {
	var err error
	f.fakeMethods["runPerfScript"].invokes++
	if f.fakeMethods["runPerfScript"].fakeReturnValues != nil && len(f.fakeMethods["runPerfScript"].fakeReturnValues) > 0 {
		f.fakeMethods["runPerfScript"].indexExecution++
		arg0 := f.fakeMethods["runPerfScript"].fakeReturnValues[f.fakeMethods["runPerfScript"].indexExecution-1].([]interface{})[0]
		if arg0 != nil {
			err = arg0.(error)
		}
	}
	return err
}

func (f *fakePerfManager) foldPerfOutput(*job.ProfilingJob, string) (error, string) {
	var err error
	var result string
	f.fakeMethods["foldPerfOutput"].invokes++
	if f.fakeMethods["foldPerfOutput"].fakeReturnValues != nil && len(f.fakeMethods["foldPerfOutput"].fakeReturnValues) > 0 {
		f.fakeMethods["foldPerfOutput"].indexExecution++
		arg0 := f.fakeMethods["foldPerfOutput"].fakeReturnValues[f.fakeMethods["foldPerfOutput"].indexExecution-1].([]interface{})[0]
		arg1 := f.fakeMethods["foldPerfOutput"].fakeReturnValues[f.fakeMethods["foldPerfOutput"].indexExecution-1].([]interface{})[1]
		if arg0 != nil {
			err = arg0.(error)
		}
		if arg1 != nil {
			result = arg1.(string)
		}
	}
	return err, result
}

func (f *fakePerfManager) handleFlamegraph(*job.ProfilingJob, flamegraph.FrameGrapher, string, string) error {
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
