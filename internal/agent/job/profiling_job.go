package job

import (
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	jsoniter "github.com/json-iterator/go"
	"time"
)

const (
	FlamegraphWidthInPixels string = "flamegraph-width-in-pixels"
)

type ProfilingJob struct {
	Duration                 time.Duration
	Interval                 time.Duration
	UID                      string
	ContainerRuntime         api.ContainerRuntime
	ContainerRuntimePath     string
	ContainerID              string
	PodUID                   string
	Language                 api.ProgrammingLanguage
	Event                    api.ProfilingEvent
	Compressor               compressor.Type
	Tool                     api.ProfilingTool
	OutputType               api.OutputType
	FileName                 string
	HeapDumpSplitInChunkSize string
	PID                      string
	Pgrep                    string
	AdditionalArguments      map[string]string
	Iteration                int
}

func (p *ProfilingJob) String() string {
	out, _ := jsoniter.MarshalToString(p)
	return out
}

func (p *ProfilingJob) ToMap() map[string]interface{} {
	out := make(map[string]interface{})
	bytes, _ := jsoniter.Marshal(p)
	_ = jsoniter.Unmarshal(bytes, &out)
	return out
}

// GetWidthAdditionalArgument returns the FlamegraphWidthInPixels argument if exists in ProfilingJob.AdditionalArguments,
// or empty string otherwise.
func (p *ProfilingJob) GetWidthAdditionalArgument() string {
	if width, ok := p.AdditionalArguments[FlamegraphWidthInPixels]; ok && stringUtils.IsNumeric(width) {
		return width
	}
	return ""
}

// DeleteWidthAdditionalArgument deletes the FlamegraphWidthInPixels argument if exists in ProfilingJob.AdditionalArguments.
func (p *ProfilingJob) DeleteWidthAdditionalArgument() {
	delete(p.AdditionalArguments, FlamegraphWidthInPixels)
}

// GetWidthAdditionalArgumentAndDelete returns the FlamegraphWidthInPixels argument if exists in ProfilingJob.AdditionalArguments
// and deletes it from the same one.
func (p *ProfilingJob) GetWidthAdditionalArgumentAndDelete() string {
	width := p.GetWidthAdditionalArgument()
	p.DeleteWidthAdditionalArgument()
	return width
}
