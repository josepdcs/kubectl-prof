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

type pprofCreator struct{}

func (p *pprofCreator) Create(targetPod *apiv1.Pod, cfg *config.ProfilerConfig) (string, *batchv1.Job, error) {
	id := string(uuid.NewUUID())
	imageName := p.getImageName(cfg.Target)

	var imagePullSecret []apiv1.LocalObjectReference
	if cfg.Target.ImagePullSecret != "" {
		imagePullSecret = []apiv1.LocalObjectReference{{Name: cfg.Target.ImagePullSecret}}
	}

	commonMeta := p.getObjectMeta(id, cfg)

	resources, err := cfg.Job.ToResourceRequirements()
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to generate resource requirements")
	}

	args := kubernetes.GetArgs(targetPod, cfg, id)
	args = append(args, "--pprof-host", targetPod.Status.PodIP)

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
					Tolerations:      cfg.Job.Tolerations,
					ImagePullSecrets: imagePullSecret,
					InitContainers:   nil,
					Containers: []apiv1.Container{
						{
							ImagePullPolicy: cfg.Target.ImagePullPolicy,
							Name:            ContainerName,
							Image:           imageName,
							Command:         []string{command},
							Args:            args,
							SecurityContext: &apiv1.SecurityContext{
								Privileged: boolPtr(false),
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
func (p *pprofCreator) getImageName(t *config.TargetConfig) string {
	var imageName string
	if t.Image != "" {
		imageName = t.Image
	} else {
		imageName = fmt.Sprintf("%s:%s-bpf", baseImageName, version.GetCurrent())
	}
	return imageName
}

func (p *pprofCreator) getObjectMeta(id string, cfg *config.ProfilerConfig) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-pprof-%s", ContainerName, id),
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
