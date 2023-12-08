package pod

import "bytes"

type ExecFake struct {
	outFake    *bytes.Buffer
	errOutFake *bytes.Buffer
	fakeError  error
}

func NewExecFake(outFake *bytes.Buffer, errOutFake *bytes.Buffer, fakeError error) *ExecFake {
	return &ExecFake{
		outFake:    outFake,
		errOutFake: errOutFake,
		fakeError:  fakeError,
	}
}

func (e *ExecFake) Execute(string, string, string, []string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	return nil, e.outFake, e.errOutFake, e.fakeError
}
