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
	"github.com/josepdcs/kubectl-perf/pkg/cli"
	"github.com/josepdcs/kubectl-perf/pkg/cli/config"
	"github.com/josepdcs/kubectl-perf/pkg/cli/handler"
	"github.com/josepdcs/kubectl-perf/pkg/cli/kubernetes"
	"log"

	v1 "k8s.io/api/core/v1"
)

type Profiler struct {
	Connector kubernetes.Connector
	Getter    kubernetes.Getter
	Creator   kubernetes.Creator
	Deleter   kubernetes.Deleter
}

//NewProfiler returns new Profiler
func NewProfiler(con kubernetes.Connector, get kubernetes.Getter, cre kubernetes.Creator, del kubernetes.Deleter) *Profiler {
	return &Profiler{
		Connector: con,
		Getter:    get,
		Creator:   cre,
		Deleter:   del,
	}
}

func (p *Profiler) Profile(cfg *config.ProfilerConfig) {
	ns, err := p.Connector.Connect(cfg.ConfigFlags)
	if err != nil {
		log.Fatalf("Failed connecting to kubernetes cluster: %v\n", err)
	}

	printer := cli.NewPrinter(cfg.Target.DryRun)

	if cfg.Target.Namespace == "" {
		cfg.Target.Namespace = ns
	}
	cfg.Job.Namespace = ns

	ctx := context.Background()

	printer.Print("Verifying target pod ... ")
	pod, err := p.Getter.GetPod(cfg.Target.PodName, cfg.Target.Namespace, ctx)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	containerName, err := validatePod(pod, cfg.Target)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	containerId, err := kubernetes.ToContainerId(containerName, pod)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	printer.PrintSuccess()

	cfg.Target.ContainerName = containerName
	cfg.Target.ContainerId = containerId

	printer.Print("Launching profiler ... ")
	profileId, job, err := p.Creator.CreateProfilingJob(pod, cfg, ctx)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	if cfg.Target.DryRun {
		return
	}

	cfg.Target.Id = profileId
	profilingPod, err := p.Getter.GetProfilingPod(cfg, ctx)
	if err != nil {
		printer.PrintError()
		log.Fatalf(err.Error())
	}

	printer.PrintSuccess()
	eventHandler := handler.NewEventHandler(job, cfg.Target, p.Deleter)
	done, err := kubernetes.GetPodLogs(profilingPod, eventHandler, ctx)
	if err != nil {
		printer.PrintError()
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
