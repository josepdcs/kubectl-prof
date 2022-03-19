package job

import (
	"fmt"
	"github.com/josepdcs/kubectl-profile/api"
	"github.com/josepdcs/kubectl-profile/pkg/cli/config"
	"github.com/josepdcs/kubectl-profile/pkg/cli/version"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type pythonCreator struct{}

func (p *pythonCreator) create(targetPod *apiv1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error) {
	id := string(uuid.NewUUID())
	var imageName string
	var imagePullSecret []apiv1.LocalObjectReference
	args := []string{
		id,
		string(targetPod.UID),
		cfg.Target.ContainerName,
		cfg.Target.ContainerId,
		cfg.Target.Duration.String(),
		string(cfg.Target.Language),
		string(cfg.Target.Event),
		string(cfg.Target.ContainerRuntime),
	}

	if cfg.Target.Pgrep != "" {
		args = append(args, cfg.Target.Pgrep)
	}

	if cfg.Target.Image != "" {
		imageName = cfg.Target.Image
	} else {
		imageName = fmt.Sprintf("%s:%s-python", baseImageName, version.GetCurrent())
	}

	if cfg.Target.ImagePullSecret != "" {
		imagePullSecret = []apiv1.LocalObjectReference{{Name: cfg.Target.ImagePullSecret}}
	}

	commonMeta := metav1.ObjectMeta{
		Name:      fmt.Sprintf("kubectl-profile-%s", id),
		Namespace: cfg.Job.Namespace,
		Labels: map[string]string{
			"kubectl-profile/id": id,
		},
		Annotations: map[string]string{
			"sidecar.istio.io/inject": "false",
		},
	}

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
					},
					ImagePullSecrets: imagePullSecret,
					InitContainers:   nil,
					Containers: []apiv1.Container{
						{
							ImagePullPolicy: apiv1.PullAlways,
							Name:            ContainerName,
							Image:           imageName,
							Command:         []string{"/app/agent"},
							Args:            args,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "target-filesystem",
									MountPath: api.GetContainerRuntimeRootPath[cfg.Target.ContainerRuntime],
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &cfg.Privileged,
								Capabilities: &apiv1.Capabilities{
									Add: cfg.Capabilities,
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
