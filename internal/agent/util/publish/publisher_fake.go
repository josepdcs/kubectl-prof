package publish

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
)

type publisherFake struct {
	retError error
}

func NewPublisherFake(retError error) Publisher {
	return &publisherFake{
		retError: retError,
	}
}

func (p publisherFake) Do(compressor.Type, string, api.OutputType) error {
	return p.retError
}

func (p publisherFake) DoWithNativeGzipAndSplit(string, string, api.OutputType) error {
	return p.retError
}
