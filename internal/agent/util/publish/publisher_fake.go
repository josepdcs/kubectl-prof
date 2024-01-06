package publish

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
)

// FakePublisher is an interface that wraps the Publisher interface
type FakePublisher interface {
	Return(fakeReturnValues ...interface{}) *Fake
	On(methodName string) *Fake
	InvokedTimes(methodName string) int

	Publisher
}

// Fake is an implementation of the FakePublisher interface
type Fake struct {
	*fakePublisher
}

// fakePublisher is a fake implementation of the Publisher interface
type fakePublisher struct {
	// values to be returned by the fake publisher
	fakeReturnValues []interface{}
	// name of the method invoked
	methodName string
	// map of method names and the number of times they were
	invokes map[string]int
}

// NewFakePublisher returns a new fake publisher
func NewFakePublisher() FakePublisher {
	return &Fake{
		fakePublisher: &fakePublisher{
			invokes: make(map[string]int),
		},
	}
}

// Return sets the values to be returned by the fake publisher
func (p *Fake) Return(fakeReturnValues ...interface{}) *Fake {
	p.fakeReturnValues = fakeReturnValues
	return p
}

// On sets the expected arguments for the fake publisher
func (p *Fake) On(methodName string) *Fake {
	p.methodName = methodName
	return p
}

// InvokedTimes represents the number of times the method was invoked
func (p *Fake) InvokedTimes(methodName string) int {
	return p.invokes[methodName]
}

// Do is a mock implementation of the Publisher interface
func (p *fakePublisher) Do(compressor.Type, string, api.OutputType) error {
	p.invokes["Do"]++
	if p.methodName == "Do" && p.fakeReturnValues != nil && len(p.fakeReturnValues) > 0 {
		return p.fakeReturnValues[0].(error)
	}
	return nil
}

// DoWithNativeGzipAndSplit is a mock implementation of the Publisher interface
func (p *fakePublisher) DoWithNativeGzipAndSplit(string, string, api.OutputType) error {
	p.invokes["DoWithNativeGzipAndSplit"]++
	if p.methodName == "DoWithNativeGzipAndSplit" && p.fakeReturnValues != nil && len(p.fakeReturnValues) > 0 {
		return p.fakeReturnValues[0].(error)
	}
	return nil
}
