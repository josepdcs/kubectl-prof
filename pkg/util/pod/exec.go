package pod

import (
	"bytes"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/exec"
)

type Exec struct {
	RestConfig    *rest.Config
	ClientSet     kubernetes.Interface
	Pod           *apiv1.Pod
	ContainerName string
}

func NewExec(config *rest.Config, client kubernetes.Interface, pod *apiv1.Pod, containerName string) *Exec {
	config.APIPath = "/api"
	config.GroupVersion = &schema.GroupVersion{Version: "v1"}
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	return &Exec{
		RestConfig:    config,
		ClientSet:     client,
		Pod:           pod,
		ContainerName: containerName,
	}
}

// ExecCmd execute command on current Exec.Pod
func (p *Exec) ExecCmd(command []string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ioStreams, in, out, errOut := genericclioptions.NewTestIOStreams()
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			Namespace:       p.Pod.Namespace,
			PodName:         p.Pod.Name,
			Stdin:           true,
			TTY:             false,
			Quiet:           false,
			InterruptParent: nil,
			IOStreams:       ioStreams,
			ContainerName:   p.ContainerName,
		},
		Command:       command,
		Executor:      &exec.DefaultRemoteExecutor{},
		PodClient:     p.ClientSet.CoreV1(),
		GetPodTimeout: 0,
		Config:        p.RestConfig,
	}

	err := options.Run()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not run exec operation: %v", err)
	}

	return in, out, errOut, nil
}
