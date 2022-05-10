package spies

import "github.com/stretchr/testify/mock"

type (
	results interface {
		Present() bool
		Get(index int) interface{}
		GetOr(int, interface{}) interface{}
		String(indexOrNil ...int) string
		Int(index int) int
		Error(index int) error
		Bool(index int) bool
	}

	missingResult struct{}

	presentResult struct {
		mock.Arguments
	}
)

func (r presentResult) Present() bool {
	return true
}

func (r presentResult) GetOr(i int, _ interface{}) interface{} {
	return r.Arguments.Get(i)
}

func (r missingResult) Present() bool {
	return false
}

func (r missingResult) GetOr(_ int, it interface{}) interface{} {
	return it
}

func (r missingResult) Get(int) interface{} {
	return nil
}

func (r missingResult) String(indexOrNil ...int) string {
	if len(indexOrNil) == 0 {
		return "a missing result"
	}
	return ""
}

func (r missingResult) Int(int) int {
	return 0
}

func (r missingResult) Error(int) error {
	return nil
}

func (r missingResult) Bool(int) bool {
	return false
}
