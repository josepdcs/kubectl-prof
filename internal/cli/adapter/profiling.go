package adapter

import (
	"bufio"
	"context"
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes/job"
	"github.com/josepdcs/kubectl-prof/internal/cli/result"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	podexec "github.com/josepdcs/kubectl-prof/pkg/util/pod"
	"io"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"path/filepath"
	"time"
)

type EventHandler interface {
	Handle(events chan string, done chan bool, resultFile chan result.File)
}

// ProfilingAdapter defines all methods related to profiling job and the target pod to be profiled
type ProfilingAdapter interface {
	// CreateProfilingJob creates the profiling job
	CreateProfilingJob(*v1.Pod, *config.ProfilerConfig, context.Context) (string, *batchv1.Job, error)
	// GetProfilingPod returns the created profiling pod from the profiling job
	GetProfilingPod(cfg *config.ProfilerConfig, ctx context.Context) (*v1.Pod, error)
	// GetProfilingPodLogs manages the logs of the profiling pod
	GetProfilingPodLogs(pod *v1.Pod, handler EventHandler, ctx context.Context) (chan bool, chan result.File, error)
	// GetRemoteFile returns the remote result file
	GetRemoteFile(pod *v1.Pod, remoteFile result.File, localPath string, c compressor.Type) (string, error)
	// DeleteProfilingJob deletes the previous created profiling job
	DeleteProfilingJob(*batchv1.Job, context.Context) error
	// GetTargetPod GetTargetPath returns the target pod to be profiled or which is being profiled
	GetTargetPod(podName, namespace string, ctx context.Context) (*v1.Pod, error)
}

// profilingAdapter implements ProfilingAdapter and wraps kubernetes.ConnectionInfo
type profilingAdapter struct {
	connectionInfo kubernetes.ConnectionInfo
}

// NewProfilingAdapter returns new instance of ProfilingAdapter
func NewProfilingAdapter(connectionInfo kubernetes.ConnectionInfo) ProfilingAdapter {
	return profilingAdapter{
		connectionInfo: connectionInfo,
	}
}

func (p profilingAdapter) GetTargetPod(podName, namespace string, ctx context.Context) (*v1.Pod, error) {
	podObject, err := p.connectionInfo.ClientSet.
		CoreV1().
		Pods(namespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return podObject, nil
}

func (p profilingAdapter) CreateProfilingJob(targetPod *v1.Pod, cfg *config.ProfilerConfig, ctx context.Context) (string, *batchv1.Job, error) {
	j, err := job.Get(cfg.Target.Language, cfg.Target.ProfilingTool)
	if err != nil {
		return "", nil, fmt.Errorf("unable to get type of job: %w", err)
	}
	id, profilingJob, err := j.Create(targetPod, cfg)
	if err != nil {
		return "", nil, fmt.Errorf("unable to create job: %w", err)
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

func (p profilingAdapter) GetProfilingPod(cfg *config.ProfilerConfig, ctx context.Context) (*v1.Pod, error) {
	var pod *v1.Pod
	err := wait.Poll(1*time.Second, 5*time.Minute,
		func() (bool, error) {
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
				return false, fmt.Errorf("pod failed")
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

func (p profilingAdapter) GetProfilingPodLogs(pod *v1.Pod, handler EventHandler, ctx context.Context) (chan bool, chan result.File, error) {
	req := p.connectionInfo.ClientSet.CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, &v1.PodLogOptions{
			Follow:    true,
			Container: job.ContainerName,
		})

	readCloser, err := req.Stream(ctx)
	if err != nil {
		return nil, nil, err
	}

	eventsChan := make(chan string)
	done := make(chan bool)
	resultFile := make(chan result.File)
	go handler.Handle(eventsChan, done, resultFile)
	go func() {
		defer func(readCloser io.ReadCloser) {
			err := readCloser.Close()
			if err != nil {
				fmt.Printf("error closing resource: %s", err)
				return
			}
		}(readCloser)

		r := bufio.NewReader(readCloser)
		for {
			bytes, err := r.ReadBytes('\n')
			if err != nil {
				return
			}
			eventsChan <- string(bytes)
		}
	}()

	return done, resultFile, nil
}

func (p profilingAdapter) GetRemoteFile(pod *v1.Pod, remoteFile result.File, localPath string, t compressor.Type) (string, error) {
	podFile := podexec.NewExec(p.connectionInfo.RestConfig, p.connectionInfo.ClientSet, pod)

	_, out, _, err := podFile.ExecCmd([]string{"/bin/cat", remoteFile.FileName})
	if err != nil {
		return "", fmt.Errorf("could not download profiler result file from pod: %w", err)
	}

	comp, err := compressor.Get(t)
	if err != nil {
		return "", fmt.Errorf("could not get compressor: %v\n", err)
	}

	decoded, err := comp.Decode(out.Bytes())
	if err != nil {
		return "", fmt.Errorf("could not decode remote file: %v\n", err)
	}

	fileName := filepath.Join(localPath, renameResultFileName(remoteFile.FileName, remoteFile.Timestamp))

	err = os.WriteFile(fileName, decoded, 0644)
	if err != nil {
		return "", fmt.Errorf("could not write result file: %w", err)
	}

	return fileName, nil
}

func renameResultFileName(fileName string, t time.Time) string {
	f := stringUtils.SubstringBeforeLast(stringUtils.SubstringAfterLast(fileName, "/"), ".")
	return stringUtils.SubstringBefore(f, ".") + "-" + t.Format(time.RFC3339) + "." + stringUtils.SubstringAfter(f, ".")
}

func (p profilingAdapter) DeleteProfilingJob(job *batchv1.Job, ctx context.Context) error {
	deleteStrategy := metav1.DeletePropagationForeground
	return p.connectionInfo.ClientSet.
		BatchV1().
		Jobs(job.Namespace).
		Delete(ctx, job.Name, metav1.DeleteOptions{
			PropagationPolicy: &deleteStrategy,
		})
}
