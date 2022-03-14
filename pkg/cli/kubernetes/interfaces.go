package kubernetes

import (
	"context"
	"github.com/josepdcs/kubectl-profile/pkg/cli/config"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type Connector interface {
	Connect(clientGetter genericclioptions.RESTClientGetter) (string, error)
}

type Getter interface {
	GetPod(podName, namespace string, ctx context.Context) (*apiv1.Pod, error)
	GetProfilingPod(cfg *config.ProfilerConfig, ctx context.Context) (*apiv1.Pod, error)
}

type Creator interface {
	CreateProfilingJob(targetPod *v1.Pod, cfg *config.ProfilerConfig, ctx context.Context) (string, *batchv1.Job, error)
}

type Deleter interface {
	DeleteProfilingJob(job *batchv1.Job, targetDetails *config.TargetConfig, ctx context.Context) error
}
