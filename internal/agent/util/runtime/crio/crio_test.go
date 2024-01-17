package crio

import (
	"github.com/josepdcs/kubectl-prof/internal/agent/testdata"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCrioConfigFile(t *testing.T) {
	assert.Equal(t, "/var/lib/containers/storage/overlay-containers/1234/userdata/config.json",
		crioConfigFile("1234", "/var/lib/containers/storage"))
}

func TestCrioStateFile(t *testing.T) {
	assert.Equal(t, "/var/lib/containers/storage/overlay-containers/1234/userdata/state.json",
		crioStateFile("1234", "/var/lib/containers/storage"))
}

func TestRootFileSystemLocation(t *testing.T) {
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
				crioConfigFile = func(string, string) string {
					return filepath.FromSlash(testdata.CrioTestDataDir() + "/other.json")
				}
			},
			containedErrMsg: "read file failed",
		},
		{
			name:                 "expect root filesystem",
			containerID:          "1234",
			containerRuntimePath: "/var/lib/containers/storage",
			mockFunc: func() {
				crioConfigFile = func(string, string) string {
					return filepath.FromSlash(testdata.CrioTestDataDir() + "/config.json")
				}
			},
			expected: "/var/lib/containers/storage/overlay/6cd0ab1d34d6895a03bc33482e9a903db973f87ce3db41176e09dc772a561052/merged",
		},
	}

	c := NewCrio()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			location, err := c.RootFileSystemLocation(tt.containerID, tt.containerRuntimePath)

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
			configFile: filepath.FromSlash(testdata.CrioTestDataDir() + "/config.json"),
			expected: "/var/lib/containers/storage/overlay/" +
				"6cd0ab1d34d6895a03bc33482e9a903db973f87ce3db41176e09dc772a561052/merged",
		},
		{
			name:            "unexpected runtime spec",
			configFile:      filepath.FromSlash(testdata.CrioTestDataDir() + "/wrong.json"),
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
			name:                 "unable read root filesystem",
			containerID:          "1234",
			containerRuntimePath: "/var/lib/containers/storage",
			mockFunc: func() {
				crioStateFile = func(string, string) string {
					return filepath.FromSlash(testdata.CrioTestDataDir() + "/other.json")
				}
			},
			containedErrMsg: "read file failed",
		},
		{
			name:                 "expect root filesystem",
			containerID:          "1234",
			containerRuntimePath: "/var/lib/containers/storage",
			mockFunc: func() {
				crioStateFile = func(string, string) string {
					return filepath.FromSlash(testdata.CrioTestDataDir() + "/state.json")
				}
			},
			expected: "3854",
		},
	}

	c := NewCrio()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			pid, err := c.PID(tt.containerID, tt.containerRuntimePath)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			}
			assert.Equal(t, tt.expected, pid)
		})
	}
}

func TestRuntimeState(t *testing.T) {
	tests := []struct {
		name            string
		configFile      string
		expected        string
		containedErrMsg string
	}{
		{
			name:       "expected runtime spec",
			configFile: filepath.FromSlash(testdata.CrioTestDataDir() + "/state.json"),
			expected:   "3854",
		},
		{
			name:            "unexpected runtime spec",
			configFile:      filepath.FromSlash(testdata.CrioTestDataDir() + "/wrong.json"),
			containedErrMsg: "unmarshal file failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result, err := runtimeState(tt.configFile)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			} else {
				assert.Equal(t, tt.expected, strconv.Itoa(result.Pid))
			}
		})
	}
}
