package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTargetConfig_DeepCopy(t *testing.T) {
	target := &TargetConfig{
		Namespace:            "namespace",
		PodName:              "pod",
		ContainerName:        "container",
		LabelSelector:        "label",
		ContainerID:          "containerID",
		Event:                "event",
		Duration:             10,
		Interval:             5,
		Id:                   "id",
		LocalPath:            "localPath",
		Alpine:               true,
		DryRun:               true,
		Image:                "image",
		ContainerRuntime:     "Containerd",
		ContainerRuntimePath: "path",
	}
	deepCopy := target.DeepCopy()
	assert.Equal(t, target, deepCopy)
}
