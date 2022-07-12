/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"bufio"
	"context"
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes/job"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	podexec "github.com/josepdcs/kubectl-prof/pkg/util/pod"
	"io"
	"io/ioutil"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

type EventHandler interface {
	Handle(events chan string, done chan bool, resultFileName chan string)
}

type getter struct {
	KubeContext
}

//NewGetter returns new implementation of Getter
func NewGetter(kubeContext KubeContext) *getter {
	return &getter{
		kubeContext,
	}
}

func (g *getter) GetPod(podName, namespace string, ctx context.Context) (*apiv1.Pod, error) {
	podObject, err := g.ClientSet.
		CoreV1().
		Pods(namespace).
		Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return podObject, nil
}

func (g *getter) GetProfilingPod(cfg *config.ProfilerConfig, ctx context.Context) (*apiv1.Pod, error) {
	var pod *apiv1.Pod
	err := wait.Poll(1*time.Second, 5*time.Minute,
		func() (bool, error) {
			podList, err := g.ClientSet.
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
			case apiv1.PodFailed:
				return false, fmt.Errorf("pod failed")
			case apiv1.PodSucceeded:
				fallthrough
			case apiv1.PodRunning:
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

func (g *getter) GetPodLogs(pod *apiv1.Pod, handler EventHandler, ctx context.Context) (chan bool, chan string, error) {
	req := g.ClientSet.CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, &apiv1.PodLogOptions{
			Follow:    true,
			Container: job.ContainerName,
		})

	readCloser, err := req.Stream(ctx)
	if err != nil {
		return nil, nil, err
	}

	eventsChan := make(chan string)
	done := make(chan bool)
	resultFileName := make(chan string)
	go handler.Handle(eventsChan, done, resultFileName)
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

	return done, resultFileName, nil
}

func (g *getter) GetRemoteFile(pod *apiv1.Pod, remoteFileName string, localFileName string, c api.Compressor) error {
	podFile := podexec.NewExec(g.RestConfig, g.ClientSet, pod)

	_, out, _, err := podFile.ExecCmd([]string{"/bin/cat", remoteFileName})
	if err != nil {
		return fmt.Errorf("could not download profiler result file from pod: %w", err)
	}

	comp, err := compressor.Get(c)
	if err != nil {
		return fmt.Errorf("could not get compressor: %v\n", err)
	}

	decoded, err := comp.Decode(out.Bytes())
	if err != nil {
		return fmt.Errorf("could not decode remote file: %v\n", err)
	}

	err = ioutil.WriteFile(localFileName, decoded, 0644)
	if err != nil {
		return fmt.Errorf("could not write result file: %w", err)
	}

	return nil
}
