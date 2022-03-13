package utils

import (
	"github.com/josepdcs/kubectl-profile/api"
	"github.com/josepdcs/kubectl-profile/pkg/agent/test"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestGetTargetFileSystemLocation(t *testing.T) {
	tests := []struct {
		name            string
		runtime         api.ContainerRuntime
		containerID     string
		mockFunc        func()
		expected        string
		containedErrMsg string
	}{
		{
			name:            "empty container runtime and container ID",
			runtime:         "",
			containerID:     "",
			mockFunc:        func() {},
			containedErrMsg: "container runtime and container ID are mandatory",
		},
		{
			name:            "empty container runtime",
			runtime:         "",
			containerID:     "1234",
			mockFunc:        func() {},
			containedErrMsg: "container runtime and container ID are mandatory",
		},
		{
			name:            "empty container ID",
			runtime:         api.Crio,
			containerID:     "",
			mockFunc:        func() {},
			containedErrMsg: "container runtime and container ID are mandatory",
		},
		{
			name:            "unexpected container runtime",
			runtime:         "other",
			containerID:     "1234",
			mockFunc:        func() {},
			containedErrMsg: "unsupported container runtime: other",
		},
		{
			name:            "expected crio runtime and unexpected container ID",
			runtime:         api.Crio,
			containerID:     "1234",
			mockFunc:        func() {},
			containedErrMsg: "read file failed:",
		},
		{
			name:        "expected crio runtime and expected container ID",
			runtime:     api.Crio,
			containerID: "1234",
			mockFunc: func() {
				crioConfigFile = func(string) string {
					return filepath.FromSlash(test.CrioTestDataDir() + "/config.json")
				}
			},
			expected: "/var/lib/containers/storage/overlay/6cd0ab1d34d6895a03bc33482e9a903db973f87ce3db41176e09dc772a561052/merged",
		},
		{
			name:            "expected containerd runtime and unexpected container ID",
			runtime:         api.Containerd,
			containerID:     "1234",
			mockFunc:        func() {},
			containedErrMsg: "containerd at not supported yet, coming soon...",
		},
		{
			name:        "expected containerd runtime and expected container ID",
			runtime:     api.Containerd,
			containerID: "1234",
			mockFunc: func() {
				crioConfigFile = func(string) string {
					return filepath.FromSlash(test.CrioTestDataDir() + "/config.json")
				}
			},
			containedErrMsg: "containerd at not supported yet, coming soon...",
		},
		{
			name:        "expected docker runtime and unexpected container ID",
			runtime:     api.Docker,
			containerID: "1234",
			mockFunc: func() {
				dockerMountIdLocation = func(string) string {
					return filepath.FromSlash(test.DockerTestDataDir() + "/other")
				}
			},
			containedErrMsg: "read file failed:",
		},
		{
			name:        "expected docker runtime and expected container ID",
			runtime:     api.Docker,
			containerID: "1234",
			mockFunc: func() {
				dockerMountIdLocation = func(string) string {
					return filepath.FromSlash(test.DockerTestDataDir() + "/mount-id")
				}
			},
			expected: "/var/lib/docker/overlay2/123456789/merged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			location, err := TargetFileSystemLocation(tt.runtime, tt.containerID)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			}
			assert.Equal(t, tt.expected, location)
		})
	}

}

func TestRuntimeSpec(t *testing.T) {
	tests := []struct {
		name            string
		configFile      string
		expected        string
		containedErrMsg string
	}{
		{
			name:       "expected runtime spec",
			configFile: filepath.FromSlash(test.CrioTestDataDir() + "/config.json"),
			expected: "/var/lib/containers/storage/overlay/" +
				"6cd0ab1d34d6895a03bc33482e9a903db973f87ce3db41176e09dc772a561052/merged",
		},
		{
			name:            "unexpected runtime spec",
			configFile:      filepath.FromSlash(test.CrioTestDataDir() + "/wrong_config.json"),
			containedErrMsg: "unmarshal file failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result, err := runtimeSpec(tt.configFile)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			} else {
				assert.Equal(t, tt.expected, result.Root.Path)
			}
		})
	}
}
