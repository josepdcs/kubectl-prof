package containerd

import (
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
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
			name:                 "expect root filesystem",
			containerID:          "1234",
			containerRuntimePath: "/run/containerd",
			mockFunc: func() {
				pidFile = func(string, string) string {
					return filepath.FromSlash(testdata.ContainerdTestDataDir() + "/init.pid")
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
