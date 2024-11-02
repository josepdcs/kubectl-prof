package kubernetes

import (
	"testing"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestToContainerId(t *testing.T) {
	type args struct {
		containerName string
		pod           *v1.Pod
	}
	tests := []struct {
		name  string
		given func() args
		when  func(args) (string, error)
		then  func(t *testing.T, result string, err error)
	}{
		{
			name: "should return ContainerID",
			given: func() args {
				return args{
					containerName: "ContainerName",
					pod: &v1.Pod{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec:       v1.PodSpec{},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Name:        "ContainerName",
									ContainerID: "ContainerID",
								},
							},
						},
					},
				}
			},
			when: func(args args) (string, error) {
				return ToContainerId(args.containerName, args.pod)
			},
			then: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				assert.Equal(t, "ContainerID", result)
			},
		},
		{
			name: "should not return ContainerID",
			given: func() args {
				return args{
					containerName: "ContainerName",
					pod: &v1.Pod{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec:       v1.PodSpec{},
						Status: v1.PodStatus{
							ContainerStatuses: []v1.ContainerStatus{
								{
									Name:        "OtherContainerName",
									ContainerID: "ContainerID",
								},
							},
						},
					},
				}
			},
			when: func(args args) (string, error) {
				return ToContainerId(args.containerName, args.pod)
			},
			then: func(t *testing.T, result string, err error) {
				require.Error(t, err)
				assert.EqualError(t, err, "Could not find container id for ContainerName")
				assert.Empty(t, result)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			args := tt.given()

			// When
			result, err := tt.when(args)

			// Then
			tt.then(t, result, err)
		})
	}
}

func TestGetArgs(t *testing.T) {
	type args struct {
		targetPod *v1.Pod
		cfg       *config.ProfilerConfig
		id        string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "With default arguments",
			args: args{
				targetPod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "UID",
						Name: "PodName",
					},
					Spec: v1.PodSpec{
						NodeName: "NodeName",
					},
				},
				cfg: &config.ProfilerConfig{
					Target: &config.TargetConfig{
						Namespace:            "",
						PodName:              "",
						ContainerName:        "",
						ContainerID:          "ContainerID",
						Event:                api.Cpu,
						Duration:             60 * time.Second,
						Id:                   "",
						LocalPath:            "LocalPath",
						DryRun:               false,
						Image:                "",
						ContainerRuntime:     api.Crio,
						ContainerRuntimePath: "/var/lib/containers/storage",
						Language:             api.Java,
						Compressor:           compressor.Gzip,
						ImagePullSecret:      "",
						ServiceAccountName:   "",
						ProfilingTool:        api.AsyncProfiler,
						OutputType:           api.FlameGraph,
						ExtraTargetOptions: config.ExtraTargetOptions{
							GracePeriodEnding: 5 * time.Minute,
						},
					},
					Job:      nil,
					LogLevel: "",
				},
				id: "ID",
			},
			want: []string{
				"--target-container-runtime", "crio",
				"--target-container-runtime-path", "/var/lib/containers/storage",
				"--target-pod-uid", "UID",
				"--target-container-id", "ContainerID",
				"--lang", "java",
				"--event-type", "cpu",
				"--compressor-type", "gzip",
				"--profiling-tool", "async-profiler",
				"--output-type", "flamegraph",
				"--grace-period-ending", "5m0s",
				"--job-id", "ID",
				"--duration", "1m0s",
			},
		},
		{
			name: "With rest of arguments",
			args: args{
				targetPod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "UID",
						Name: "PodName",
					},
					Spec: v1.PodSpec{
						NodeName: "NodeName",
					},
				},
				cfg: &config.ProfilerConfig{
					Target: &config.TargetConfig{
						Namespace:            "",
						PodName:              "",
						ContainerName:        "",
						ContainerID:          "ContainerID",
						Event:                api.Cpu,
						Duration:             60 * time.Second,
						Interval:             10 * time.Second,
						Id:                   "",
						LocalPath:            "LocalPath",
						DryRun:               false,
						Image:                "",
						ContainerRuntime:     api.Crio,
						ContainerRuntimePath: "/var/lib/containers/storage",
						Language:             api.Java,
						Compressor:           compressor.Gzip,
						ImagePullSecret:      "",
						ServiceAccountName:   "",
						ProfilingTool:        api.Jcmd,
						OutputType:           api.HeapDump,
						ExtraTargetOptions: config.ExtraTargetOptions{
							PrintLogs:                true,
							GracePeriodEnding:        5 * time.Minute,
							HeapDumpSplitInChunkSize: "10M",
						},
					},
					Job:      nil,
					LogLevel: "",
				},
				id: "ID",
			},
			want: []string{
				"--target-container-runtime", "crio",
				"--target-container-runtime-path", "/var/lib/containers/storage",
				"--target-pod-uid", "UID",
				"--target-container-id", "ContainerID",
				"--lang", "java",
				"--event-type", "cpu",
				"--compressor-type", "gzip",
				"--profiling-tool", "jcmd",
				"--output-type", "heapdump",
				"--grace-period-ending", "5m0s",
				"--job-id", "ID",
				"--duration", "1m0s",
				"--interval", "10s",
				"--print-logs",
				"--heap-dump-split-in-chunk-size", "10M",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetArgs(tt.args.targetPod, tt.args.cfg, tt.args.id), "getArgs(%v, %v, %v)", tt.args.targetPod, tt.args.cfg, tt.args.id)
		})
	}
}
