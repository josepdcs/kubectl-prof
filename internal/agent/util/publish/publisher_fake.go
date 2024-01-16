package publish

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
)

// FakePublisher is an interface that wraps the Publisher interface
type FakePublisher interface {
	// On sets the expected method to be invoked by the fake publisher
	On(methodName string) *FakeMethod

	Publisher
}

// Fake is an implementation of the FakePublisher interface
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

// NewFakePublisher returns a new FakePublisher
func NewFakePublisher() FakePublisher {
	return &Fake{
		fakeMethods: make(map[string]*FakeMethod),
	}
}

// On sets the expected method to be invoked by the fake publisher
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

// Do is a fake implementation of the Publisher interface
func (f *Fake) Do(compressor.Type, string, api.OutputType) error {
	f.fakeMethods["Do"].invokes++
	if f.fakeMethods["Do"].fakeReturnValues != nil && len(f.fakeMethods["Do"].fakeReturnValues) > 0 {
		f.fakeMethods["Do"].indexExecution++
		if f.fakeMethods["Do"].fakeReturnValues[f.fakeMethods["Do"].indexExecution-1].([]interface{})[0] != nil {
			return f.fakeMethods["Do"].fakeReturnValues[f.fakeMethods["Do"].indexExecution-1].([]interface{})[0].(error)
		}
	}
	return nil
}

// DoWithNativeGzipAndSplit is a fake implementation of the Publisher interface
func (f *Fake) DoWithNativeGzipAndSplit(string, string, api.OutputType) error {
	f.fakeMethods["DoWithNativeGzipAndSplit"].invokes++
	if f.fakeMethods["DoWithNativeGzipAndSplit"].fakeReturnValues != nil && len(f.fakeMethods["DoWithNativeGzipAndSplit"].fakeReturnValues) > 0 {
		f.fakeMethods["DoWithNativeGzipAndSplit"].indexExecution++
		if f.fakeMethods["DoWithNativeGzipAndSplit"].fakeReturnValues[f.fakeMethods["DoWithNativeGzipAndSplit"].indexExecution-1].([]interface{})[0] != nil {
			return f.fakeMethods["DoWithNativeGzipAndSplit"].fakeReturnValues[f.fakeMethods["DoWithNativeGzipAndSplit"].indexExecution-1].([]interface{})[0].(error)
		}
	}
	return nil
}
