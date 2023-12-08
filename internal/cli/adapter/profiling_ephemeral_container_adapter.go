package adapter

import (
	"context"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes/ephemeral"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	kuberrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"time"
)

// ProfilingEphemeralContainerAdapter defines all methods related to ephemeral container for profiling
type ProfilingEphemeralContainerAdapter interface {
	// AddEphemeralContainer creates the ephemeral container for profiling and adds it to the target pod
	AddEphemeralContainer(*v1.Pod, *config.ProfilerConfig, context.Context, time.Duration) (*v1.Pod, error)
	// GetEphemeralContainerName returns the container name of the current ephemeral container
	GetEphemeralContainerName() string
}

// profilingEphemeralContainerAdapter implements ProfilingEphemeralContainerAdapter and wraps kubernetes.ConnectionInfo
type profilingEphemeralContainerAdapter struct {
	connectionInfo kubernetes.ConnectionInfo

	containerName string
}

// NewProfilingEphemeralContainerAdapter returns new instance of ProfilingEphemeralContainerAdapter
func NewProfilingEphemeralContainerAdapter(connectionInfo kubernetes.ConnectionInfo) ProfilingEphemeralContainerAdapter {
	return &profilingEphemeralContainerAdapter{
		connectionInfo: connectionInfo,
	}
}

func (p *profilingEphemeralContainerAdapter) AddEphemeralContainer(targetPod *v1.Pod, cfg *config.ProfilerConfig, ctx context.Context, timeout time.Duration) (*v1.Pod, error) {
	j, err := ephemeral.Get(cfg.Target.Language)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get the ephemeral container type")
	}
	ephemeralContainer := j.Create(targetPod, cfg)
	p.containerName = ephemeralContainer.Name

	// target pod is copied and the ephemeral container added to the copied one
	copied := targetPod.DeepCopy()
	copied.Spec.EphemeralContainers = append(copied.Spec.EphemeralContainers, *ephemeralContainer)

	if cfg.Target.DryRun {
		err = printPod(copied)
		return nil, err
	}

	targetPodJS, err := jsoniter.Marshal(targetPod)
	if err != nil {
		return nil, errors.Wrap(err, "error creating JSON for original target pod")
	}

	copiedJS, err := jsoniter.Marshal(copied)
	if err != nil {
		return nil, errors.Wrap(err, "error creating JSON for copied pod: %v")
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(targetPodJS, copiedJS, targetPod)
	if err != nil {
		return nil, errors.Wrap(err, "error creating patch to add ephemeral container")
	}

	patchedPod, err := p.connectionInfo.ClientSet.
		CoreV1().
		Pods(targetPod.GetNamespace()).
		Patch(ctx, targetPod.GetName(), types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "ephemeralcontainers")

	if err != nil {
		var serr *kuberrors.StatusError
		if errors.As(err, &serr) && serr.Status().Reason == metav1.StatusReasonNotFound && serr.ErrStatus.Details.Name == "" {
			return nil, errors.Wrap(err, "ephemeral containers are disabled for this cluster")
		}

		return nil, err
	}

	return p.waitForEphemeralContainer(patchedPod, ctx, timeout)
}

func printPod(pod *v1.Pod) error {
	encoder := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml: true,
	})

	return encoder.Encode(pod, os.Stdout)
}

func (p *profilingEphemeralContainerAdapter) waitForEphemeralContainer(targetPod *v1.Pod, ctx context.Context, timeout time.Duration) (*v1.Pod, error) {
	var pod *v1.Pod
	err := wait.Poll(1*time.Second, timeout,
		func() (bool, error) {
			podList, err := p.connectionInfo.ClientSet.
				CoreV1().
				Pods(targetPod.GetNamespace()).
				List(ctx, metav1.ListOptions{
					FieldSelector: fields.OneTermEqualSelector("metadata.name", targetPod.GetName()).String(),
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
				return false, errors.Errorf("target pod now is failed: %s", targetPod.GetName())
			case v1.PodRunning:
				s := getEphemeralContainerStatusByName(pod, p.GetEphemeralContainerName())
				if s != nil && (s.State.Running != nil || s.State.Terminated != nil) {
					return true, nil
				}
				// Keep watching
				return false, nil
			default:
				return false, nil
			}
		})

	if err != nil {
		return nil, err
	}

	return pod, nil
}

func getEphemeralContainerStatusByName(pod *v1.Pod, containerName string) *v1.ContainerStatus {
	allContainerStatus := [][]v1.ContainerStatus{pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses, pod.Status.EphemeralContainerStatuses}
	for _, statusSlice := range allContainerStatus {
		for i := range statusSlice {
			if statusSlice[i].Name == containerName {
				return &statusSlice[i]
			}
		}
	}
	return nil
}

func (p *profilingEphemeralContainerAdapter) GetEphemeralContainerName() string {
	return p.containerName
}
