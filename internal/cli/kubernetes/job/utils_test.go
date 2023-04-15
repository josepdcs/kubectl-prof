package job

import (
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func Test_getArgs(t *testing.T) {
	type args struct {
		targetPod *apiv1.Pod
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
				targetPod: &apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "UID",
						Name: "PodName",
					},
					Spec: apiv1.PodSpec{
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
						ContainerRuntimePath: "",
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
				"--job-id", "ID",
				"--target-container-runtime", "crio",
				"--target-pod-uid", "UID",
				"--target-container-id", "ContainerID",
				"--lang", "java",
				"--event-type", "cpu",
				"--compressor-type", "gzip",
				"--profiling-tool", "async-profiler",
				"--output-type", "flamegraph",
				"--grace-period-ending", "5m0s",
				"--duration", "1m0s",
			},
		},
		{
			name: "With rest of arguments",
			args: args{
				targetPod: &apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						UID:  "UID",
						Name: "PodName",
					},
					Spec: apiv1.PodSpec{
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
						ContainerRuntimePath: "",
						Language:             api.Java,
						Compressor:           compressor.Gzip,
						ImagePullSecret:      "",
						ServiceAccountName:   "",
						ProfilingTool:        api.AsyncProfiler,
						OutputType:           api.FlameGraph,
						ExtraTargetOptions: config.ExtraTargetOptions{
							PrintLogs:         true,
							GracePeriodEnding: 5 * time.Minute,
						},
					},
					Job:      nil,
					LogLevel: "",
				},
				id: "ID",
			},
			want: []string{
				"--job-id", "ID",
				"--target-container-runtime", "crio",
				"--target-pod-uid", "UID",
				"--target-container-id", "ContainerID",
				"--lang", "java",
				"--event-type", "cpu",
				"--compressor-type", "gzip",
				"--profiling-tool", "async-profiler",
				"--output-type", "flamegraph",
				"--grace-period-ending", "5m0s",
				"--duration", "1m0s",
				"--print-logs",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getArgs(tt.args.targetPod, tt.args.cfg, tt.args.id), "getArgs(%v, %v, %v)", tt.args.targetPod, tt.args.cfg, tt.args.id)
		})
	}
}
