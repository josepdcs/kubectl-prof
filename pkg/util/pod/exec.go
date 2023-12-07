package pod

import (
	"bytes"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/exec"
)

// Executor interface for execute command on pod
type Executor interface {
	Execute(namespace, podName, containerName string, command []string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error)
}

// Exec struct for execute command on pod
type Exec struct {
	RestConfig *rest.Config
	ClientSet  kubernetes.Interface
}

// NewExec create new Exec
func NewExec(config *rest.Config, client kubernetes.Interface) *Exec {
	config.APIPath = "/api"
	config.GroupVersion = &schema.GroupVersion{Version: "v1"}
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	return &Exec{
		RestConfig: config,
		ClientSet:  client,
	}
}

// Execute execute command on current podName and containerName
func (p *Exec) Execute(namespace, podName, containerName string, command []string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ioStreams, in, out, errOut := genericiooptions.NewTestIOStreams()
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			Namespace:       namespace,
			PodName:         podName,
			Stdin:           true,
			TTY:             false,
			Quiet:           false,
			InterruptParent: nil,
			IOStreams:       ioStreams,
			ContainerName:   containerName,
		},
		Command:       command,
		Executor:      &exec.DefaultRemoteExecutor{},
		PodClient:     p.ClientSet.CoreV1(),
		GetPodTimeout: 0,
		Config:        p.RestConfig,
	}

	err := options.Run()
	if err != nil {
		return in, out, errOut, fmt.Errorf("could not run exec operation: %v", err)
	}

	return in, out, errOut, nil
}
