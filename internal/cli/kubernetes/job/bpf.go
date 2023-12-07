package job

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/version"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type bpfCreator struct{}

func (b *bpfCreator) Create(targetPod *apiv1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error) {
	id := string(uuid.NewUUID())
	imageName := b.getImageName(cfg.Target)

	var imagePullSecret []apiv1.LocalObjectReference
	if cfg.Target.ImagePullSecret != "" {
		imagePullSecret = []apiv1.LocalObjectReference{{Name: cfg.Target.ImagePullSecret}}
	}

	commonMeta := b.getObjectMeta(id, cfg)

	resources, err := cfg.Job.ToResourceRequirements()
	if err != nil {
		return "", nil, fmt.Errorf("unable to generate resource requirements: %w", err)
	}

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "JobConfig",
			APIVersion: "batch/v1",
		},
		ObjectMeta: commonMeta,
		Spec: batchv1.JobSpec{
			Parallelism:             int32Ptr(1),
			Completions:             int32Ptr(1),
			TTLSecondsAfterFinished: int32Ptr(5),
			BackoffLimit:            int32Ptr(2),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: commonMeta,
				Spec: apiv1.PodSpec{
					HostPID: true,
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
					ImagePullSecrets: imagePullSecret,
					InitContainers:   nil,
					Containers: []apiv1.Container{
						{
							ImagePullPolicy: cfg.Target.ImagePullPolicy,
							Name:            ContainerName,
							Image:           imageName,
							Command:         []string{command},
							Args:            kubernetes.GetArgs(targetPod, cfg, id),
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "target-filesystem",
									MountPath: api.GetContainerRuntimeRootPath[cfg.Target.ContainerRuntime],
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
					RestartPolicy: "Never",
					NodeName:      targetPod.Spec.NodeName,
				},
			},
		},
	}

	if cfg.Target.ServiceAccountName != "" {
		job.Spec.Template.Spec.ServiceAccountName = cfg.Target.ServiceAccountName
	}

	return id, job, nil
}

// getImageName if image name is provider from config.TargetConfig this one is returned otherwise a new one is built
func (b *bpfCreator) getImageName(t *config.TargetConfig) string {
	var imageName string
	if t.Image != "" {
		imageName = t.Image
	} else {
		imageName = fmt.Sprintf("%s:%s-bpf", baseImageName, version.GetCurrent())
	}
	return imageName
}

func (b *bpfCreator) getObjectMeta(id string, cfg *config.ProfilerConfig) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s-bpf-%s", ContainerName, cfg.Target.Language, id),
		Namespace: cfg.Job.Namespace,
		Labels: map[string]string{
			LabelID: id,
		},
		Annotations: map[string]string{
			"sidecar.istio.io/inject": "false",
			"linkerd.io/inject":       "disabled",
		},
	}
}
