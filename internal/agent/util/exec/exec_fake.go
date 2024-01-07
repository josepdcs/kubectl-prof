package exec

import (
	"os/exec"
)

// FakeCommander is an interface that wraps the Commander interface
type FakeCommander interface {
	// On sets the expected method to be invoked by the fake commander
	On(methodName string) *FakeMethod

	Commander
}

// Fake is an implementation of the FakeCommander interface
type Fake struct {
	fakeMethods map[string]*FakeMethod
}

// FakeMethod wraps the fakeMethod struct
type FakeMethod struct {
	*fakeMethod
}

// fakeMethod represents a fake method
type fakeMethod struct {
	// values to be returned by the fake method
	fakeReturnValues []interface{}
	// number of times the method was invoked
	invokes int
	// index of the execution of the method
	indexExecution int
}

// NewFakeCommander returns a new FakeCommander
func NewFakeCommander() FakeCommander {
	return &Fake{
		fakeMethods: make(map[string]*FakeMethod),
	}
}

// On sets the expected method to be invoked by the fake commander
func (f *Fake) On(methodName string) *FakeMethod {
	if _, ok := f.fakeMethods[methodName]; !ok {
		f.fakeMethods[methodName] = &FakeMethod{&fakeMethod{}}
	}
	if f.fakeMethods[methodName].fakeReturnValues == nil {
		f.fakeMethods[methodName].fakeReturnValues = make([]interface{}, 0)
	}
	return f.fakeMethods[methodName]
}

// Return sets the values to be returned by the fake method
func (f *FakeMethod) Return(fakeReturnValues ...interface{}) *FakeMethod {
	f.fakeReturnValues = append(f.fakeReturnValues, fakeReturnValues)
	return f
}

// InvokedTimes represents the number of times the method was invoked
func (f *FakeMethod) InvokedTimes() int {
	return f.invokes
}

// Command is a fake implementation of the Commander interface
func (f *Fake) Command(string, ...string) *exec.Cmd {
	f.fakeMethods["Command"].invokes++
	if f.fakeMethods["Command"].fakeReturnValues != nil && len(f.fakeMethods["Command"].fakeReturnValues) > 0 {
		f.fakeMethods["Command"].indexExecution++
		if f.fakeMethods["Command"].fakeReturnValues[f.fakeMethods["Command"].indexExecution-1].([]interface{})[0] != nil {
			return f.fakeMethods["Command"].fakeReturnValues[f.fakeMethods["Command"].indexExecution-1].([]interface{})[0].(*exec.Cmd)
		}
	}
	return nil
}

// Execute is a fake implementation of the Commander interface
func (f *Fake) Execute(*exec.Cmd) (int, []byte, error) {
	var exitCode = 0
	var output []byte
	var err error
	f.fakeMethods["Execute"].invokes++
	if f.fakeMethods["Execute"].fakeReturnValues != nil && len(f.fakeMethods["Execute"].fakeReturnValues) > 0 {
		f.fakeMethods["Execute"].indexExecution++
		if f.fakeMethods["Execute"].fakeReturnValues[f.fakeMethods["Execute"].indexExecution-1].([]interface{})[0] != nil {
			exitCode = f.fakeMethods["Execute"].fakeReturnValues[f.fakeMethods["Execute"].indexExecution-1].([]interface{})[0].(int)
		}
		if f.fakeMethods["Execute"].fakeReturnValues[f.fakeMethods["Execute"].indexExecution-1].([]interface{})[1] != nil {
			output = f.fakeMethods["Execute"].fakeReturnValues[f.fakeMethods["Execute"].indexExecution-1].([]interface{})[1].([]byte)
		}
		if f.fakeMethods["Execute"].fakeReturnValues[f.fakeMethods["Execute"].indexExecution-1].([]interface{})[2] != nil {
			err = f.fakeMethods["Execute"].fakeReturnValues[f.fakeMethods["Execute"].indexExecution-1].([]interface{})[2].(error)
		}
	}
	return exitCode, output, err
}
