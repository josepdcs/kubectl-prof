package job

import (
	"testing"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_bpfCreate_create(t *testing.T) {
	targetPod := &apiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID: "UID",
		},
		Spec: apiv1.PodSpec{
			NodeName: "NodeName",
		},
	}
	cfg := &config.ProfilerConfig{
		Target: &config.TargetConfig{
			Namespace:            "Namespace",
			PodName:              "PodName",
			ContainerName:        "ContainerName",
			ContainerID:          "ContainerID",
			Event:                "Event",
			Duration:             100,
			Id:                   "ID",
			LocalPath:            "LocalPath",
			Alpine:               false,
			DryRun:               false,
			Image:                "Image",
			ContainerRuntime:     "ContainerRuntime",
			ContainerRuntimePath: "ContainerRuntimePath",
			Language:             "Language",
			Compressor:           "Compressor",
			ImagePullSecret:      "ImagePullSecret",
			ServiceAccountName:   "ServiceAccountName",
			ImagePullPolicy:      apiv1.PullAlways,
		},
		Job: &config.JobConfig{
			ContainerConfig: config.ContainerConfig{
				RequestConfig: config.ResourceConfig{
					CPU:    "100m",
					Memory: "100Mi",
				},
				LimitConfig: config.ResourceConfig{
					CPU:    "200m",
					Memory: "200Mi",
				},
				Privileged: false,
			},
			Namespace: "Namespace",
		},
	}
	b := &bpfCreator{}
	id, job, err := b.Create(targetPod, cfg)

	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, job)

	wantedObjectMeta := b.getObjectMeta(id, cfg)
	assert.Equal(t, job.ObjectMeta, wantedObjectMeta)

	resources, err := cfg.Job.ToResourceRequirements()

	wantedJob := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "JobConfig",
			APIVersion: "batch/v1",
		},
		ObjectMeta: wantedObjectMeta,
		Spec: batchv1.JobSpec{
			Parallelism:             int32Ptr(1),
			Completions:             int32Ptr(1),
			TTLSecondsAfterFinished: int32Ptr(5),
			BackoffLimit:            int32Ptr(2),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: wantedObjectMeta,
				Spec: apiv1.PodSpec{
					HostPID:     true,
					Tolerations: cfg.Job.Tolerations,
					Volumes: []apiv1.Volume{
						{
							Name: "target-filesystem",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: cfg.Target.ContainerRuntimePath,
								},
							},
						},
						{
							Name: "modules",
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
					},
					ImagePullSecrets: []apiv1.LocalObjectReference{{Name: cfg.Target.ImagePullSecret}},
					InitContainers:   nil,
					Containers: []apiv1.Container{
						{
							ImagePullPolicy: apiv1.PullAlways,
							Name:            ContainerName,
							Image:           cfg.Target.Image,
							Command:         []string{command},
							Args:            kubernetes.GetArgs(targetPod, cfg, id),
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "target-filesystem",
									MountPath: cfg.Target.ContainerRuntimePath,
								},
								{
									Name:      "modules",
									MountPath: "/lib/modules",
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &cfg.Job.Privileged,
								Capabilities: &apiv1.Capabilities{
									Add: []apiv1.Capability{"SYS_ADMIN"},
								},
							},
							Resources: resources,
						},
					},
					RestartPolicy:      "Never",
					NodeName:           targetPod.Spec.NodeName,
					ServiceAccountName: cfg.Target.ServiceAccountName,
				},
			},
		},
	}

	assert.Equal(t, job, wantedJob)
}

func Test_bpfCreate_shouldFailWhenUnableGenerateResources(t *testing.T) {
	targetPod := &apiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID: "UID",
		},
		Spec: apiv1.PodSpec{
			NodeName: "NodeName",
		},
	}
	cfg := &config.ProfilerConfig{
		Target: &config.TargetConfig{
			Namespace:            "Namespace",
			PodName:              "PodName",
			ContainerName:        "ContainerName",
			ContainerID:          "ContainerID",
			Event:                "Event",
			Duration:             100,
			Id:                   "ID",
			LocalPath:            "LocalPath",
			DryRun:               false,
			Image:                "Image",
			ContainerRuntime:     "ContainerRuntime",
			ContainerRuntimePath: "ContainerRuntimePath",
			Language:             "Language",
			Compressor:           "Compressor",
			ServiceAccountName:   "ServiceAccountName",
		},
		Job: &config.JobConfig{
			ContainerConfig: config.ContainerConfig{
				RequestConfig: config.ResourceConfig{
					CPU:    "error",
					Memory: "100Mi",
				},
				LimitConfig: config.ResourceConfig{
					CPU:    "error",
					Memory: "200Mi",
				},
				Privileged: false,
			},
			Namespace: "Namespace",
		},
	}
	b := &bpfCreator{}
	id, job, err := b.Create(targetPod, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to generate resource requirements")
	assert.Empty(t, id)
	assert.Empty(t, job)

}

func Test_bpfCreator_getImageName(t *testing.T) {
	type args struct {
		cfg *config.TargetConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Get image name from TargetConfig",
			args: args{
				cfg: &config.TargetConfig{
					Image: "Image",
				},
			},
			want: "Image",
		},
		{
			name: "Get default image",
			args: args{
				cfg: &config.TargetConfig{
					Image: "",
				},
			},
			want: "josepdcs/kubectl-prof:-bpf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bpfCreator{}
			assert.Equalf(t, tt.want, b.getImageName(tt.args.cfg), "getImageName(%v)", tt.args.cfg)
		})
	}
}
