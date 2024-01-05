package publish

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
)

type PublisherFake struct {
	DoInvokedTimes                       int
	DoWithNativeGzipAndSplitInvokedTimes int

	// private fields
	fakeReturnValues []interface{}
}

func NewPublisherFake() Publisher {
	return &PublisherFake{}
}

func (p *PublisherFake) WithFakeReturnValues(fakeReturnValues []interface{}) *PublisherFake {
	p.fakeReturnValues = fakeReturnValues
	return p
}

func (p *PublisherFake) Do(compressor.Type, string, api.OutputType) error {
	p.DoInvokedTimes++
	if p.fakeReturnValues != nil && len(p.fakeReturnValues) > 0 {
		return p.fakeReturnValues[0].(error)
	}
	return nil
}

func (p *PublisherFake) DoWithNativeGzipAndSplit(string, string, api.OutputType) error {
	p.DoWithNativeGzipAndSplitInvokedTimes++
	if p.fakeReturnValues != nil && len(p.fakeReturnValues) > 0 {
		return p.fakeReturnValues[0].(error)
	}
	return nil
}
