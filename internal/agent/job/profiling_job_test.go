package job

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProfilingJob_String(t *testing.T) {
	p := ProfilingJob{
		Duration:         10,
		Interval:         5,
		UID:              "ID",
		ContainerRuntime: "ContainerRuntime",
		ContainerID:      "ContainerID",
		PodUID:           "PodUID",
		Language:         "Language",
		Event:            "Event",
		Compressor:       "Compressor",
		Tool:             "Tool",
		OutputType:       "OutputType",
		FileName:         "FileName",
	}

	out := p.String()

	assert.NotEmpty(t, out)
	assert.Contains(t, out, "10")
	assert.Contains(t, out, "5")
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "ContainerRuntime")
	assert.Contains(t, out, "ContainerID")
	assert.Contains(t, out, "PodUID")
	assert.Contains(t, out, "Language")
	assert.Contains(t, out, "Event")
	assert.Contains(t, out, "Compressor")
	assert.Contains(t, out, "Tool")
	assert.Contains(t, out, "FileName")
}

func TestProfilingJob_ToMap(t *testing.T) {
	p := ProfilingJob{
		Duration:         10,
		Interval:         5,
		UID:              "ID",
		ContainerRuntime: "ContainerRuntime",
		ContainerID:      "ContainerID",
		PodUID:           "PodUID",
		Language:         "Language",
		Event:            "Event",
		Compressor:       "Compressor",
		Tool:             "Tool",
		OutputType:       "OutputType",
		FileName:         "FileName",
	}

	out := p.ToMap()

	assert.NotEmpty(t, out)
	assert.Len(t, out, 12)
	assert.Equal(t, float64(10), out["Duration"])
	assert.Equal(t, float64(5), out["Interval"])
	assert.Equal(t, "ID", out["UID"])
	assert.Equal(t, "ContainerRuntime", out["ContainerRuntime"])
	assert.Equal(t, "ContainerID", out["ContainerID"])
	assert.Equal(t, "PodUID", out["PodUID"])
	assert.Equal(t, "Language", out["Language"])
	assert.Equal(t, "Event", out["Event"])
	assert.Equal(t, "Compressor", out["Compressor"])
	assert.Equal(t, "Tool", out["Tool"])
	assert.Equal(t, "OutputType", out["OutputType"])
	assert.Equal(t, "FileName", out["FileName"])
}
