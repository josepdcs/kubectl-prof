package api

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes/job"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/wait"
)

// ProfilingJobApi defines all methods related to profiling job and the target pod to be profiled
type ProfilingJobApi interface {
	// CreateProfilingJob creates the profiling job
	CreateProfilingJob(*v1.Pod, *config.ProfilerConfig, context.Context) (string, *batchv1.Job, error)
	// GetProfilingPod returns the created profiling pod from the profiling job
	GetProfilingPod(*config.ProfilerConfig, context.Context, time.Duration) (*v1.Pod, error)
	// GetProfilingContainerName returns the container name of the profiling pod
	GetProfilingContainerName() string
	// DeleteProfilingJob deletes the previous created profiling job
	DeleteProfilingJob(*batchv1.Job, context.Context) error
}

// profilingJobApi implements ProfilingJobApi and wraps kubernetes.ConnectionInfo
type profilingJobApi struct {
	connectionInfo kubernetes.ConnectionInfo
}

// NewProfilingJobApi returns new instance of ProfilingJobApi
func NewProfilingJobApi(connectionInfo kubernetes.ConnectionInfo) ProfilingJobApi {
	return &profilingJobApi{
		connectionInfo: connectionInfo,
	}
}

func (p *profilingJobApi) CreateProfilingJob(targetPod *v1.Pod, cfg *config.ProfilerConfig, ctx context.Context) (string, *batchv1.Job, error) {
	j, err := job.Get(cfg.Target.Language, cfg.Target.ProfilingTool)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to get the job type")
	}
	id, profilingJob, err := j.Create(targetPod, cfg)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to create job")
	}

	if cfg.Target.DryRun {
		err = printJob(profilingJob)
		return "", nil, err
	}
	createJob, err := p.connectionInfo.ClientSet.
		BatchV1().
		Jobs(cfg.Job.Namespace).
		Create(ctx, profilingJob, metav1.CreateOptions{})
	if err != nil {
		return "", nil, err
	}

	return id, createJob, nil
}

func printJob(job *batchv1.Job) error {
	encoder := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml: true,
	})

	return encoder.Encode(job, os.Stdout)
}

func (p *profilingJobApi) GetProfilingPod(cfg *config.ProfilerConfig, ctx context.Context, timeout time.Duration) (*v1.Pod, error) {
	var pod *v1.Pod
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := wait.PollUntilContextTimeout(ctx, 1*time.Second, timeout, true,
		func(ctx context.Context) (bool, error) {
			podList, err := p.connectionInfo.ClientSet.
				CoreV1().
				Pods(cfg.Job.Namespace).
				List(ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("%s=%s", job.LabelID, cfg.Target.Id),
				})

			if err != nil {
				return false, err
			}

			if len(podList.Items) == 0 {
				return false, nil
			}

			pod = &podList.Items[0]
			switch pod.Status.Phase {
			case v1.PodFailed:
				return false, errors.New("profiling pod failed")
			case v1.PodSucceeded:
				fallthrough
			case v1.PodRunning:
				return true, nil
			default:
				return false, nil
			}
		})

	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (p *profilingJobApi) GetProfilingContainerName() string {
	return job.ContainerName
}

func (p *profilingJobApi) DeleteProfilingJob(job *batchv1.Job, ctx context.Context) error {
	deleteStrategy := metav1.DeletePropagationForeground
	return p.connectionInfo.ClientSet.
		BatchV1().
		Jobs(job.Namespace).
		Delete(ctx, job.Name, metav1.DeleteOptions{
			PropagationPolicy: &deleteStrategy,
		})
}
