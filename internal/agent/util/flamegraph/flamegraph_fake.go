package flamegraph

import "errors"

type FlameGrapherFake struct {
}

func NewFlameGrapherFake() *FlameGrapherFake {
	return &FlameGrapherFake{}
}

func (*FlameGrapherFake) StackSamplesToFlameGraph(inputFileName string, outputFileName string) error {
	return nil
}

type FlameGrapherFakeWithError struct {
}

func NewFlameGrapherFakeWithError() *FlameGrapherFakeWithError {
	return &FlameGrapherFakeWithError{}
}

func (*FlameGrapherFakeWithError) StackSamplesToFlameGraph(inputFileName string, outputFileName string) error {
	return errors.New("StackSamplesToFlameGraph with error")
}
