package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProfilingJob_String(t *testing.T) {
	p := ProfilingJob{
		Duration:          0,
		ID:                "",
		ContainerRuntime:  "",
		ContainerID:       "",
		ContainerName:     "",
		PodUID:            "",
		Language:          "",
		TargetProcessName: "",
		Event:             "",
		Compressor:        "",
		ProfilingTool:     "",
		OutputType:        "",
		FileName:          "",
	}

	assert.NotEmpty(t, p.String())
}
