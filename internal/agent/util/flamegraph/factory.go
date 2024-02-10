package flamegraph

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"strings"
)

// FrameGrapher is an interface for converting stacks samples to flame graphs
type FrameGrapher interface {
	StackSamplesToFlameGraph(inputFileName string, outputFileName string) error
}

// Get returns an instance of FrameGrapher
func Get(job *job.ProfilingJob) FrameGrapher {
	language := strings.ToTitle(string(job.Language))
	title := fmt.Sprintf("%s - CPU Flamegraph", language)
	switch job.Language {
	case api.Python, api.Go:
		return NewFlameGrapherScript(
			WithTitle(title),
			WithWidth(job.GetWidthAdditionalArgumentAndDelete()),
		)
	case api.Node:
		return NewFlameGrapherScript(
			WithTitle(title),
			WithWidth(job.GetWidthAdditionalArgumentAndDelete()),
			WithColors("js"),
		)
	case api.Clang, api.ClangPlusPlus:
		return NewFlameGrapherScript(
			WithTitle(title),
			WithWidth(job.GetWidthAdditionalArgumentAndDelete()),
			WithColors("mem"),
		)
	case api.FakeLang:
		return NewFlameGrapherFake()
	}

	// for tests purpose
	return NewFlameGrapherFakeWithError()
}
