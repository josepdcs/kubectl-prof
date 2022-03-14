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

package profiler

import (
	"context"
	"errors"
	"fmt"
	"github.com/josepdcs/kubectl-profile/pkg/cli"
	"github.com/josepdcs/kubectl-profile/pkg/cli/config"
	"github.com/josepdcs/kubectl-profile/pkg/cli/handler"
	"github.com/josepdcs/kubectl-profile/pkg/cli/kubernetes"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"log"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

type KubeConnector interface {
	Connect(clientGetter genericclioptions.RESTClientGetter) (string, error)
}

//NewKubeConnector returns new KubeConnector
func NewKubeConnector() *kubernetes.KubeConnector {
	return &kubernetes.KubeConnector{}
}

type KubeGetter interface {
	GetPod(podName, namespace string, ctx context.Context) (*apiv1.Pod, error)
	GetProfilingPod(cfg *config.ProfileConfig, ctx context.Context) (*apiv1.Pod, error)
}

func NewKubeGetter() *kubernetes.KubeGetter {
	return &kubernetes.KubeGetter{}
}

type Profiler struct {
	KubeConnecter KubeConnector
	KubeGetter    KubeGetter
}

//NewProfiler returns new Profiler
func NewProfiler(connector KubeConnector, getter KubeGetter) *Profiler {
	return &Profiler{
		KubeConnecter: connector,
		KubeGetter:    getter,
	}
}

func (pf *Profiler) Profile(cfg *config.ProfileConfig) {
	ns, err := pf.KubeConnecter.Connect(cfg.ConfigFlags)
	if err != nil {
		log.Fatalf("Failed connecting to kubernetes cluster: %v\n", err)
	}

	p := cli.NewPrinter(cfg.Target.DryRun)

	if cfg.Target.Namespace == "" {
		cfg.Target.Namespace = ns
	}
	cfg.Job.Namespace = ns

	ctx := context.Background()

	p.Print("Verifying target pod ... ")
	pod, err := pf.KubeGetter.GetPod(cfg.Target.PodName, cfg.Target.Namespace, ctx)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	containerName, err := validatePod(pod, cfg.Target)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	containerId, err := kubernetes.ToContainerId(containerName, pod)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	p.PrintSuccess()

	cfg.Target.ContainerName = containerName
	cfg.Target.ContainerId = containerId

	p.Print("Launching profiler ... ")
	profileId, job, err := kubernetes.LaunchProfilingJob(pod, cfg, ctx)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	if cfg.Target.DryRun {
		return
	}

	cfg.Target.Id = profileId
	profilerPod, err := pf.KubeGetter.GetProfilingPod(cfg, ctx)
	if err != nil {
		p.PrintError()
		log.Fatalf(err.Error())
	}

	p.PrintSuccess()
	apiHandler := handler.NewApiEventsHandler(job, cfg.Target)
	done, err := kubernetes.GetPodLogs(profilerPod, apiHandler, ctx)
	if err != nil {
		p.PrintError()
		fmt.Println(err.Error())
	}

	<-done
}

func validatePod(pod *v1.Pod, cfg *config.TargetConfig) (string, error) {
	if pod == nil {
		return "", errors.New(fmt.Sprintf("Could not find pod %s in Namespace %s",
			cfg.PodName, cfg.Namespace))
	}

	if len(pod.Spec.Containers) != 1 {
		var containerNames []string
		for _, container := range pod.Spec.Containers {
			if container.Name == cfg.ContainerName {
				return container.Name, nil // Found given container
			}

			containerNames = append(containerNames, container.Name)
		}

		return "", errors.New(fmt.Sprintf("Could not determine container. please specify one of %v",
			containerNames))
	}

	return pod.Spec.Containers[0].Name, nil
}
