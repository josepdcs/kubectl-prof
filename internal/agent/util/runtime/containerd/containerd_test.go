package containerd

import (
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestRootFileSystemLocation(t *testing.T) {
	tests := []struct {
		name            string
		containerID     string
		expected        string
		containedErrMsg string
	}{
		{
			name:            "empty container ID",
			containedErrMsg: "container ID is mandatory",
		},
		{
			name:        "expect root filesystem",
			containerID: "1234",
			expected:    "/run/containerd/io.containerd.runtime.v2.task/k8s.io/1234/rootfs",
		},
	}

	c := NewContainerd()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := c.RootFileSystemLocation(tt.containerID)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			}
			assert.Equal(t, tt.expected, location)
		})
	}
}

func TestPidFileFromContainerID(t *testing.T) {
	assert.Equal(t, "/run/containerd/io.containerd.runtime.v2.task/k8s.io/1234/init.pid", pidFile("1234"))
}

func TestPID(t *testing.T) {
	tests := []struct {
		name            string
		containerID     string
		mockFunc        func()
		expected        string
		containedErrMsg string
	}{
		{
			name:            "empty container ID",
			mockFunc:        func() {},
			containedErrMsg: "container ID is mandatory",
		},
		{
			name:        "unable to read pid file",
			containerID: "1234",
			mockFunc: func() {
				pidFile = func(string) string {
					return filepath.FromSlash(testdata.ContainerdTestDataDir() + "/other")
				}
			},
			containedErrMsg: "no such file or directory",
		},
		{
			name:        "expect root filesystem",
			containerID: "1234",
			mockFunc: func() {
				pidFile = func(string) string {
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
			location, err := c.PID(tt.containerID)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			}
			assert.Equal(t, tt.expected, location)
		})
	}
}
