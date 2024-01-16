package jvm

import (
	"bytes"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"time"
)

// FakeJcmdManager is an interface that wraps the JcmdManager interface
type FakeJcmdManager interface {
	// On sets the expected method to be invoked by the fake JcmdManager
	On(methodName string) *fakeJcmdManagerMethod

	JcmdManager
}

// fakeJcmdManager is an implementation of the FakeJcmdManager interface
type fakeJcmdManager struct {
	fakeMethods map[string]*fakeJcmdManagerMethod
}

// fakeMethod represents a fake method
type fakeJcmdManagerMethod struct {
	// values to be returned by the fake method
	fakeReturnValues []interface{}
	// number of times the method was invoked
	invokes int
	// index of the execution of the method
	indexExecution int
}

// newFakeJcmdManager returns a new fake jcmd manager
func newFakeJcmdManager() FakeJcmdManager {
	return &fakeJcmdManager{
		fakeMethods: make(map[string]*fakeJcmdManagerMethod),
	}
}

// On sets the expected method to be invoked by the fake publisher
func (f *fakeJcmdManager) On(methodName string) *fakeJcmdManagerMethod {
	if _, ok := f.fakeMethods[methodName]; !ok {
		f.fakeMethods[methodName] = &fakeJcmdManagerMethod{}
	}
	if f.fakeMethods[methodName].fakeReturnValues == nil {
		f.fakeMethods[methodName].fakeReturnValues = make([]interface{}, 0)
	}
	return f.fakeMethods[methodName]
}

// Return sets the values to be returned by the fake method
func (f *fakeJcmdManagerMethod) Return(fakeReturnValues ...interface{}) *fakeJcmdManagerMethod {
	f.fakeReturnValues = append(f.fakeReturnValues, fakeReturnValues)
	return f
}

// InvokedTimes represents the number of times the method was invoked
func (f *fakeJcmdManagerMethod) InvokedTimes() int {
	return f.invokes
}

func (f *fakeJcmdManager) invoke(*job.ProfilingJob, string) (error, time.Duration) {
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

func (f *fakeJcmdManager) removeTmpDir() error {
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

func (f *fakeJcmdManager) linkTmpDirToTargetTmpDir(s string) error {
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

func (f *fakeJcmdManager) cleanUp(profilingJob *job.ProfilingJob, s string) {
	f.fakeMethods["cleanUp"].invokes++
	if f.fakeMethods["cleanUp"].fakeReturnValues != nil && len(f.fakeMethods["cleanUp"].fakeReturnValues) > 0 {
		f.fakeMethods["cleanUp"].indexExecution++
	}
}

func (f *fakeJcmdManager) copyJfrSettingsToTmpDir() error {
	var err error
	f.fakeMethods["copyJfrSettingsToTmpDir"].invokes++
	if f.fakeMethods["copyJfrSettingsToTmpDir"].fakeReturnValues != nil && len(f.fakeMethods["copyJfrSettingsToTmpDir"].fakeReturnValues) > 0 {
		f.fakeMethods["copyJfrSettingsToTmpDir"].indexExecution++
		arg0 := f.fakeMethods["copyJfrSettingsToTmpDir"].fakeReturnValues[f.fakeMethods["copyJfrSettingsToTmpDir"].indexExecution-1].([]interface{})[0]
		if arg0 != nil {
			err = arg0.(error)
		}
	}
	return err
}

func (f *fakeJcmdManager) handleProfilingResult(job *job.ProfilingJob, fileName string, out bytes.Buffer, targetPID string) error {
	var err error
	f.fakeMethods["handleProfilingResult"].invokes++
	if f.fakeMethods["handleProfilingResult"].fakeReturnValues != nil && len(f.fakeMethods["handleProfilingResult"].fakeReturnValues) > 0 {
		f.fakeMethods["handleProfilingResult"].indexExecution++
		arg0 := f.fakeMethods["handleProfilingResult"].fakeReturnValues[f.fakeMethods["handleProfilingResult"].indexExecution-1].([]interface{})[0]
		if arg0 != nil {
			err = arg0.(error)
		}
	}
	return err
}

func (f *fakeJcmdManager) handleJcmdRecording(targetPID string, outputType string) {
	f.fakeMethods["handleJcmdRecording"].invokes++
	if f.fakeMethods["handleJcmdRecording"].fakeReturnValues != nil && len(f.fakeMethods["handleJcmdRecording"].fakeReturnValues) > 0 {
		f.fakeMethods["handleJcmdRecording"].indexExecution++
	}
}

func (f *fakeJcmdManager) publishResult(compressor compressor.Type, fileName string, outputType api.OutputType, heapDumpSplitInChunkSize string) error {
	var err error
	f.fakeMethods["publishResult"].invokes++
	if f.fakeMethods["publishResult"].fakeReturnValues != nil && len(f.fakeMethods["publishResult"].fakeReturnValues) > 0 {
		f.fakeMethods["publishResult"].indexExecution++
		arg0 := f.fakeMethods["publishResult"].fakeReturnValues[f.fakeMethods["publishResult"].indexExecution-1].([]interface{})[0]
		if arg0 != nil {
			err = arg0.(error)
		}
	}
	return err
}
