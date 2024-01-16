package util

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/containerd"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/crio"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/runtime/fake"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

func TestNormalizeContainerID(t *testing.T) {
	tests := []string{
		"cri-o://b3f6972fb3a9f5d1eba91e43900b0839aad99f7428d0580ba1b4e501017ee949",
		"containerd://b3f6972fb3a9f5d1eba91e43900b0839aad99f7428d0580ba1b4e501017ee949",
	}
	for _, tc := range tests {
		result := NormalizeContainerID(tc)
		assert.Equal(t, "b3f6972fb3a9f5d1eba91e43900b0839aad99f7428d0580ba1b4e501017ee949", result)
	}
}

func TestContainerFileSystem(t *testing.T) {
	tests := []struct {
		name            string
		runtime         api.ContainerRuntime
		containerID     string
		mockFunc        func()
		wanted          string
		containedErrMsg string
	}{
		{
			name:            "empty container runtime",
			containerID:     "ID",
			mockFunc:        func() {},
			containedErrMsg: "container runtime and container ID are mandatory",
		},
		{
			name:            "empty container ID",
			runtime:         api.ContainerRuntime("fake"),
			mockFunc:        func() {},
			containedErrMsg: "container runtime and container ID are mandatory",
		},
		{
			name:            "unknown container runtime",
			runtime:         api.ContainerRuntime("other"),
			containerID:     "1234",
			mockFunc:        func() {},
			containedErrMsg: "unsupported container runtime: other",
		},
		{
			name:        "crio container runtime",
			runtime:     api.Crio,
			containerID: "1234",
			mockFunc: func() {
				runtime = func(runtime api.ContainerRuntime) (Container, error) {
					return fake.NewRuntimeFake(), nil
				}
			},
			wanted: "/root/fs/1234",
		},
		{
			name:        "containerd container runtime",
			runtime:     api.Containerd,
			containerID: "1234",
			mockFunc: func() {
				runtime = func(runtime api.ContainerRuntime) (Container, error) {
					return fake.NewRuntimeFake(), nil
				}
			},
			wanted: "/root/fs/1234",
		},
	}

	// preserve the original function
	original := runtime
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			location, err := ContainerFileSystem(tt.runtime, tt.containerID)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			}
			assert.Equal(t, tt.wanted, location)
		})
	}
	runtime = original
}

func TestRuntime(t *testing.T) {
	tests := []struct {
		name            string
		runtime         api.ContainerRuntime
		expected        Container
		containedErrMsg string
	}{
		{
			name:            "empty container runtime",
			containedErrMsg: "container runtime is are mandatory",
		},
		{
			name:            "unknown container runtime",
			runtime:         api.ContainerRuntime("other"),
			containedErrMsg: "unsupported container runtime: other",
		},
		{
			name:     "crio container runtime",
			runtime:  api.Crio,
			expected: crio.NewCrio(),
		},
		{
			name:     "containerd container runtime",
			runtime:  api.Containerd,
			expected: containerd.NewContainerd(),
		},
		{
			name:     "fake container runtime",
			runtime:  api.FakeContainer,
			expected: fake.NewRuntimeFake(),
		},
		{
			name:     "fake container runtime with RootFileSystemLocationResultError",
			runtime:  api.FakeContainerWithRootFileSystemLocationResultError,
			expected: fake.NewRuntimeFake().WithRootFileSystemLocationResultError(),
		},
		{
			name:     "fake container runtime with PIDResultError",
			runtime:  api.FakeContainerWithPIDResultError,
			expected: fake.NewRuntimeFake().WithPIDResultError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := runtime(tt.runtime)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			} else {
				assert.IsType(t, tt.expected, r)
			}
		})
	}
}

type childPIDGetterMock struct {
	results    []string
	interation int
}

// getChildPID returns the PID of the child process of the given PID
func (c *childPIDGetterMock) get(string) string {
	defer func() { c.interation++ }()
	if c.interation >= len(c.results) {
		return ""
	}
	return c.results[c.interation]
}

func TestContainerPID(t *testing.T) {
	tests := []struct {
		name            string
		job             job.ProfilingJob
		mockFunc        func()
		expected        string
		containedErrMsg string
	}{
		{
			name: "empty container runtime",
			job: job.ProfilingJob{
				ContainerID: "1234",
			},
			mockFunc:        func() {},
			containedErrMsg: "container runtime and container ID are mandatory",
		},
		{
			name: "empty container ID",
			job: job.ProfilingJob{
				ContainerRuntime: api.ContainerRuntime("other"),
				ContainerID:      "",
			},
			mockFunc:        func() {},
			containedErrMsg: "container runtime and container ID are mandatory",
		},
		{
			name: "unknown container runtime",
			job: job.ProfilingJob{
				ContainerRuntime: api.ContainerRuntime("other"),
				ContainerID:      "1234",
			},
			mockFunc:        func() {},
			containedErrMsg: "unsupported container runtime: other",
		},
		{
			name: "crio container runtime",
			job: job.ProfilingJob{
				ContainerRuntime: api.Crio,
				ContainerID:      "12334_CRIO",
			},
			mockFunc: func() {
				runtime = func(runtime api.ContainerRuntime) (Container, error) {
					return fake.NewRuntimeFake(), nil
				}
				childPIDGetterInstance = &childPIDGetterMock{
					interation: 0,
					results:    []string{"PID_12334_CRIO"},
				}
			},
			expected: "PID_12334_CRIO",
		},
		{
			name: "containerd container runtime",
			job: job.ProfilingJob{
				ContainerRuntime: api.Containerd,
				ContainerID:      "12334_CONTAINERD",
			},
			mockFunc: func() {
				runtime = func(runtime api.ContainerRuntime) (Container, error) {
					return fake.NewRuntimeFake(), nil
				}
				childPIDGetterInstance = &childPIDGetterMock{
					interation: 0,
					results:    []string{"PID_12334_CONTAINERD"},
				}
			},
			expected: "PID_12334_CONTAINERD",
		},
		{
			name: "more than one child process",
			job: job.ProfilingJob{
				ContainerRuntime: api.Containerd,
				ContainerID:      "12334_CONTAINERD",
			},
			mockFunc: func() {
				runtime = func(runtime api.ContainerRuntime) (Container, error) {
					return fake.NewRuntimeFake(), nil
				}
				childPIDGetterInstance = &childPIDGetterMock{
					interation: 0,
					results:    []string{"PID_12334_CONTAINERD\nPID_12335_CONTAINERD"},
				}
			},
			expected: "PID_12334_CONTAINERD",
		},
	}

	// preserve the original function
	original := runtime
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			pid, err := ContainerPID(&tt.job)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
			}
			assert.Equal(t, tt.expected, pid)
		})
	}
	runtime = original
}

func TestGetCandidatePIDs(t *testing.T) {
	tests := []struct {
		name            string
		job             job.ProfilingJob
		mockFunc        func()
		expected        []string
		containedErrMsg string
	}{
		{
			name: "empty container runtime",
			job: job.ProfilingJob{
				ContainerID: "1234",
			},
			mockFunc:        func() {},
			expected:        nil,
			containedErrMsg: "container runtime and container ID are mandatory",
		},
		{
			name: "empty container ID",
			job: job.ProfilingJob{
				ContainerRuntime: api.ContainerRuntime("other"),
				ContainerID:      "",
			},
			mockFunc:        func() {},
			expected:        nil,
			containedErrMsg: "container runtime and container ID are mandatory",
		},
		{
			name: "unknown container runtime",
			job: job.ProfilingJob{
				ContainerRuntime: api.ContainerRuntime("other"),
				ContainerID:      "1234",
			},
			mockFunc:        func() {},
			expected:        nil,
			containedErrMsg: "unsupported container runtime: other",
		},
		{
			name: "crio container runtime",
			job: job.ProfilingJob{
				ContainerRuntime: api.Crio,
				ContainerID:      "12334_CRIO",
			},
			mockFunc: func() {
				runtime = func(runtime api.ContainerRuntime) (Container, error) {
					return fake.NewRuntimeFake(), nil
				}
				childPIDGetterInstance = &childPIDGetterMock{
					interation: 0,
					results:    []string{"PID_12334_CRIO"},
				}
			},
			expected: []string{"PID_12334_CRIO"},
		},
		{
			name: "containerd container runtime",
			job: job.ProfilingJob{
				ContainerRuntime: api.Containerd,
				ContainerID:      "12334_CONTAINERD",
			},
			mockFunc: func() {
				runtime = func(runtime api.ContainerRuntime) (Container, error) {
					return fake.NewRuntimeFake(), nil
				}
				childPIDGetterInstance = &childPIDGetterMock{
					interation: 0,
					results:    []string{"PID_12334_CONTAINERD"},
				}
			},
			expected: []string{"PID_12334_CONTAINERD"},
		},
		{
			name: "more than one child process",
			job: job.ProfilingJob{
				ContainerRuntime: api.Containerd,
				ContainerID:      "12334_CONTAINERD",
			},
			mockFunc: func() {
				runtime = func(runtime api.ContainerRuntime) (Container, error) {
					return fake.NewRuntimeFake(), nil
				}
				childPIDGetterInstance = &childPIDGetterMock{
					interation: 0,
					results:    []string{"PID_12334_CONTAINERD\nPID_12335_CONTAINERD"},
				}
			},
			expected: []string{"PID_12334_CONTAINERD", "PID_12335_CONTAINERD"},
		},
	}

	// preserve the original function
	original := runtime
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()
			pids, err := GetCandidatePIDs(&tt.job)

			if err != nil {
				assert.Contains(t, err.Error(), tt.containedErrMsg)
				assert.Nil(t, pids)
			} else {
				assert.Equal(t, tt.expected, pids)
			}
		})
	}
	runtime = original
}

func Test_childPIDGetter_get(t *testing.T) {
	type args struct {
		pid string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "pid not exists",
			args: args{
				pid: "1234",
			},
			want: "",
		},
		{
			name: "pid exists",
			args: args{
				pid: strconv.Itoa(os.Getpid()),
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := childPIDGetter{}
			assert.Equalf(t, tt.want, c.get(tt.args.pid), "get(%v)", tt.args.pid)
		})
	}
}
