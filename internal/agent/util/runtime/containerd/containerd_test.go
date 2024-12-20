package containerd

import (
	"path/filepath"
	"testing"

	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	"github.com/stretchr/testify/assert"
)

func TestRootFileSystemLocation(t *testing.T) {
	tests := []struct {
		name                 string
		containerID          string
		containerRuntimePath string
		expected             string
		containedErrMsg      string
	}{
		{
			name:            "empty container ID",
			containedErrMsg: "container ID is mandatory",
		},
		{
			name:            "empty container runtime path",
			containerID:     "1234",
			containedErrMsg: "container runtime path is mandatory",
		},
		{
			name:                 "expect root filesystem",
			containerID:          "1234",
			containerRuntimePath: "/run/containerd",
			expected:             "/run/containerd/io.containerd.runtime.v2.task/k8s.io/1234/rootfs",
		},
	}

	c := NewContainerd()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := c.RootFileSystemLocation(tt.containerID, tt.containerRuntimePath)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			}
			assert.Equal(t, tt.expected, location)
		})
	}
}

func TestPidFileFromContainerID(t *testing.T) {
	assert.Equal(t, "/run/containerd/io.containerd.runtime.v2.task/k8s.io/1234/init.pid", pidFile("1234", "/run/containerd"))
	assert.Equal(t, "/run/containerd/io.containerd.runtime.v2.task/k8s.io/1234/1234.pid", pidContainerIDFile("1234", "/run/containerd"))
}

func TestConfigFile(t *testing.T) {
	assert.Equal(t, "/run/containerd/io.containerd.runtime.v2.task/k8s.io/1234/config.json", configFile("1234", "/run/containerd"))
}

func TestPID(t *testing.T) {
	tests := []struct {
		name                 string
		containerID          string
		containerRuntimePath string
		mockFunc             func()
		expected             string
		containedErrMsg      string
	}{
		{
			name:            "empty container ID",
			mockFunc:        func() {},
			containedErrMsg: "container ID is mandatory",
		},
		{
			name:            "empty container runtime path",
			containerID:     "1234",
			mockFunc:        func() {},
			containedErrMsg: "container runtime path is mandatory",
		},
		{
			name:                 "unable to read pid file",
			containerID:          "1234",
			containerRuntimePath: "/run/containerd",
			mockFunc: func() {
				pidFile = func(string, string) string {
					return filepath.FromSlash(testdata.ContainerdTestDataDir() + "/other")
				}
			},
			containedErrMsg: "no such file or directory",
		},
		{
			name:                 "expect pid file",
			containerID:          "1234",
			containerRuntimePath: "/run/containerd",
			mockFunc: func() {
				pidFile = func(string, string) string {
					return filepath.FromSlash(testdata.ContainerdTestDataDir() + "/init.pid")
				}
			},
			expected: "123456",
		},
		{
			name:                 "expect pid file with container ID",
			containerID:          "1234",
			containerRuntimePath: "/run/containerd",
			mockFunc: func() {
				pidFile = func(string, string) string {
					return filepath.FromSlash(testdata.ContainerdTestDataDir() + "/1234.pid")
				}
			},
			expected: "123456",
		},
	}

	c := NewContainerd()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			location, err := c.PID(tt.containerID, tt.containerRuntimePath)

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
			configFile: filepath.FromSlash(testdata.ContainerdTestDataDir() + "/config.json"),
			expected: "/var/lib/containers/storage/overlay/" +
				"6cd0ab1d34d6895a03bc33482e9a903db973f87ce3db41176e09dc772a561052/merged",
		},
		{
			name:            "unexpected runtime spec",
			configFile:      filepath.FromSlash(testdata.ContainerdTestDataDir() + "/wrong.json"),
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

func TestCWD(t *testing.T) {
	tests := []struct {
		name                 string
		containerID          string
		containerRuntimePath string
		mockFunc             func()
		expected             string
		containedErrMsg      string
	}{
		{
			name:            "empty container ID",
			mockFunc:        func() {},
			containedErrMsg: "container ID is mandatory",
		},
		{
			name:            "empty container runtime path",
			containerID:     "1234",
			mockFunc:        func() {},
			containedErrMsg: "container runtime path is mandatory",
		},
		{
			name:                 "unable read root filesystem",
			containerID:          "1234",
			containerRuntimePath: "/var/lib/containers/storage",
			mockFunc: func() {
				configFile = func(string, string) string {
					return filepath.FromSlash(testdata.CrioTestDataDir() + "/other.json")
				}
			},
			containedErrMsg: "read file failed",
		},
		{
			name:                 "expect cwd",
			containerID:          "1234",
			containerRuntimePath: "/var/lib/containers/storage",
			mockFunc: func() {
				configFile = func(string, string) string {
					return filepath.FromSlash(testdata.CrioTestDataDir() + "/config.json")
				}
			},
			expected: "/opt/app",
		},
	}

	c := NewContainerd()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			location, err := c.CWD(tt.containerID, tt.containerRuntimePath)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			}
			assert.Equal(t, tt.expected, location)
		})
	}
}
