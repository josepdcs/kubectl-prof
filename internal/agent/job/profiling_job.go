package job

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	jsoniter "github.com/json-iterator/go"
	"time"
)

type ProfilingJob struct {
	Duration         time.Duration
	Interval         time.Duration
	UID              string
	ContainerRuntime api.ContainerRuntime
	ContainerID      string
	PodUID           string
	Language         api.ProgrammingLanguage
	Event            api.ProfilingEvent
	Compressor       compressor.Type
	Tool             api.ProfilingTool
	OutputType       api.OutputType
	FileName         string
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
