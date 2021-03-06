package kubernetes

import (
	"context"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/pkg/cli/config"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

type Connector interface {
	Connect(clientGetter genericclioptions.RESTClientGetter) (kubernetes.Interface, string, error)
}

type Getter interface {
	GetPod(podName, namespace string, ctx context.Context) (*apiv1.Pod, error)
	GetProfilingPod(cfg *config.ProfilerConfig, ctx context.Context) (*apiv1.Pod, error)
	GetPodLogs(pod *apiv1.Pod, handler EventHandler, ctx context.Context) (chan bool, chan string, error)
	GetRemoteFile(pod *apiv1.Pod, remoteFileName string, localFileName string, compressor api.Compressor) error
}

type Creator interface {
	CreateProfilingJob(*v1.Pod, *config.ProfilerConfig, context.Context) (string, *batchv1.Job, error)
}

type Deleter interface {
	DeleteProfilingJob(*batchv1.Job, context.Context) error
}
