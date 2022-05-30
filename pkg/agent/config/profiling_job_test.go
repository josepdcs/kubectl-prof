package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProfilingJob_String(t *testing.T) {
	p := ProfilingJob{
		Duration:          10,
		ID:                "ID",
		ContainerRuntime:  "ContainerRuntime",
		ContainerID:       "ContainerID",
		ContainerName:     "ContainerName",
		PodUID:            "PodUID",
		Language:          "Language",
		TargetProcessName: "TargetProcessName",
		Event:             "Event",
		Compressor:        "Compressor",
		ProfilingTool:     "ProfilingTool",
		OutputType:        "OutputType",
		FileName:          "FileName",
	}

	out := p.String()

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "10")
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "ContainerRuntime")
	assert.Contains(t, out, "ContainerID")
	assert.Contains(t, out, "ContainerName")
	assert.Contains(t, out, "PodUID")
	assert.Contains(t, out, "Language")
	assert.Contains(t, out, "TargetProcessName")
	assert.Contains(t, out, "Event")
	assert.Contains(t, out, "Compressor")
	assert.Contains(t, out, "ProfilingTool")
	assert.Contains(t, out, "FileName")
}

func TestProfilingJob_ToMap(t *testing.T) {
	p := ProfilingJob{
		Duration:          10,
		ID:                "ID",
		ContainerRuntime:  "ContainerRuntime",
		ContainerID:       "ContainerID",
		ContainerName:     "ContainerName",
		PodUID:            "PodUID",
		Language:          "Language",
		TargetProcessName: "TargetProcessName",
		Event:             "Event",
		Compressor:        "Compressor",
		ProfilingTool:     "ProfilingTool",
		OutputType:        "OutputType",
		FileName:          "FileName",
	}

	out := p.ToMap()

	assert.NotEmpty(t, out)
	assert.Len(t, out, 13)
	assert.Equal(t, float64(10), out["Duration"])
	assert.Equal(t, "ID", out["ID"])
	assert.Equal(t, "ContainerRuntime", out["ContainerRuntime"])
	assert.Equal(t, "ContainerID", out["ContainerID"])
	assert.Equal(t, "ContainerName", out["ContainerName"])
	assert.Equal(t, "PodUID", out["PodUID"])
	assert.Equal(t, "Language", out["Language"])
	assert.Equal(t, "TargetProcessName", out["TargetProcessName"])
	assert.Equal(t, "Event", out["Event"])
	assert.Equal(t, "Compressor", out["Compressor"])
	assert.Equal(t, "ProfilingTool", out["ProfilingTool"])
	assert.Equal(t, "OutputType", out["OutputType"])
	assert.Equal(t, "FileName", out["FileName"])
}
