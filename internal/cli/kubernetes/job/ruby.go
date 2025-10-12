package job

import (
	"fmt"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/version"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type rubyCreator struct{}

var rubyDefaultCapabilities = []apiv1.Capability{"SYS_PTRACE"}

func (r *rubyCreator) Create(targetPod *apiv1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error) {
	id := string(uuid.NewUUID())
	imageName := r.getImageName(cfg.Target)

	var imagePullSecret []apiv1.LocalObjectReference
	if cfg.Target.ImagePullSecret != "" {
		imagePullSecret = []apiv1.LocalObjectReference{{Name: cfg.Target.ImagePullSecret}}
	}

	capabilities := cfg.Job.Capabilities
	if len(capabilities) == 0 {
		capabilities = rubyDefaultCapabilities
	}

	commonMeta := r.getObjectMeta(id, cfg)

	resources, err := cfg.Job.ToResourceRequirements()
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to generate resource requirements")
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
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: commonMeta,
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
					},
					ImagePullSecrets: imagePullSecret,
					InitContainers:   nil,
					Containers: []apiv1.Container{
						{
							ImagePullPolicy: apiv1.PullAlways,
							Name:            ContainerName,
							Image:           imageName,
							Command:         []string{command},
							Args:            kubernetes.GetArgs(targetPod, cfg, id),
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "target-filesystem",
									MountPath: cfg.Target.ContainerRuntimePath,
								},
							},
							SecurityContext: &apiv1.SecurityContext{
								Privileged: &cfg.Job.Privileged,
								Capabilities: &apiv1.Capabilities{
									Add: capabilities,
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
func (r *rubyCreator) getImageName(t *config.TargetConfig) string {
	var imageName string
	if t.Image != "" {
		imageName = t.Image
	} else {
		imageName = fmt.Sprintf("%s:%s-ruby", baseImageName, version.GetCurrent())
	}
	return imageName
}

func (r *rubyCreator) getObjectMeta(id string, cfg *config.ProfilerConfig) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-ruby-%s", ContainerName, id),
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
