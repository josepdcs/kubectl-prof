package flamegraph

import "errors"

type FlameGrapherFake struct {
	StackSamplesToFlameGraphInvoked bool
}

func NewFlameGrapherFake() *FlameGrapherFake {
	return &FlameGrapherFake{}
}

func (f *FlameGrapherFake) StackSamplesToFlameGraph(inputFileName string, outputFileName string) error {
	f.StackSamplesToFlameGraphInvoked = true
	return nil
}

type FlameGrapherFakeWithError struct {
	StackSamplesToFlameGraphInvoked bool
}

func NewFlameGrapherFakeWithError() *FlameGrapherFakeWithError {
	return &FlameGrapherFakeWithError{}
}

func (f *FlameGrapherFakeWithError) StackSamplesToFlameGraph(inputFileName string, outputFileName string) error {
	f.StackSamplesToFlameGraphInvoked = true
	return errors.New("StackSamplesToFlameGraph with error")
}
