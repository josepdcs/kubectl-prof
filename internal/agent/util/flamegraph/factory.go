package flamegraph

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
)

// FrameGrapher is an interface for converting stacks samples to flame graphs
type FrameGrapher interface {
	StackSamplesToFlameGraph(inputFileName string, outputFileName string) error
}

// Get returns an instance of FrameGrapher
func Get(job *job.ProfilingJob) FrameGrapher {
	switch job.Language {
	case api.Python:
		return NewFlameGrapherScript(
			WithTitle("Python - CPU Flamegraph"),
			WithWidth(job.GetWidthAdditionalArgumentAndDelete()),
		)
	case api.Go:
		return NewFlameGrapherScript(
			WithTitle("Golang - CPU Flamegraph"),
			WithWidth(job.GetWidthAdditionalArgumentAndDelete()),
		)
	case api.Node:
		return NewFlameGrapherScript(
			WithTitle("NodeJS - CPU Flamegraph"),
			WithWidth(job.GetWidthAdditionalArgumentAndDelete()),
			WithColors("js"),
		)
	case api.FakeLang:
		return NewFlameGrapherFake()
	}

	// for tests purpose
	return NewFlameGrapherFakeWithError()
}
