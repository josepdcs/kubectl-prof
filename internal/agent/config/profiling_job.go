package config

import (
	"github.com/josepdcs/kubectl-prof/api"
	jsoniter "github.com/json-iterator/go"
	"time"
)

type ProfilingJob struct {
	Duration          time.Duration
	ID                string
	ContainerRuntime  api.ContainerRuntime
	ContainerID       string
	ContainerName     string
	PodUID            string
	Language          api.ProgrammingLanguage
	TargetProcessName string
	Event             api.ProfilingEvent
	Compressor        api.Compressor
	ProfilingTool     api.ProfilingTool
	OutputType        api.EventType
	FileName          string
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
