package pod

import (
	"bytes"
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/cp"
	"k8s.io/kubectl/pkg/cmd/exec"
	"os"
)

type PodExec struct {
	RestConfig *rest.Config
	*kubernetes.Clientset
	Namespace     string
	PodName       string
	ContainerName string
}

func NewPodExec(config rest.Config, clientset *kubernetes.Clientset, namespace, podname, containername string) *PodExec {
	config.APIPath = "/api"
	config.GroupVersion = &schema.GroupVersion{Version: "v1"}
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	return &PodExec{
		RestConfig:    &config,
		Clientset:     clientset,
		Namespace:     namespace,
		PodName:       podname,
		ContainerName: containername,
	}
}

// ExecCmd
// EXAMPLES
// // Execute ls -l /tmp on container test-container on pod test in namespace default
// // and print the resulting output
// in, out, errOut, err := podExec.ExecCmd([]string{"ls", "-l", "/tmp"}, "default", "test", "test-container")
// if err != nil {
//		fmt.Printf("%v\n", err)
//	}
//	fmt.Println("out:")
//	fmt.Printf("%s", out.String()) // will execute ls -l /tmp in the pod and output the result
func (p *PodExec) ExecCmd(command []string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ioStreams, in, out, errOut := genericclioptions.NewTestIOStreams()
	options := &exec.ExecOptions{
		StreamOptions: exec.StreamOptions{
			Namespace:       p.Namespace,
			PodName:         p.PodName,
			ContainerName:   p.ContainerName,
			Stdin:           true,
			TTY:             false,
			Quiet:           false,
			InterruptParent: nil,
			IOStreams:       ioStreams,
		},
		Command:       command,
		Executor:      &exec.DefaultRemoteExecutor{},
		PodClient:     p.Clientset.CoreV1(),
		GetPodTimeout: 0,
		Config:        p.RestConfig,
	}

	err := options.Run()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Could not run exec operation: %v", err)
	}

	return in, out, errOut, nil
}

func (p *PodExec) UploadFileTwo(config *rest.Config, clientset *kubernetes.Clientset, path string) ([]byte, error) {
	options := &exec.ExecOptions{}
	out := bytes.NewBuffer([]byte{})
	errOut := bytes.NewBuffer([]byte{})
	f, _ := os.Open(path)
	reader, writer := io.Pipe()

	go func(f *os.File, writer *io.PipeWriter) {
		defer writer.Close()
		_, _ = io.Copy(writer, f)
	}(f, writer)

	options.StreamOptions = exec.StreamOptions{
		IOStreams: genericclioptions.IOStreams{
			In:     reader,
			Out:    out,
			ErrOut: errOut,
		},
		Stdin:     true,
		Namespace: "default",
		PodName:   "cron-test",
	}
	options.Executor = &exec.DefaultRemoteExecutor{}
	options.Namespace = "default"
	options.PodName = "cron-test"
	options.ContainerName = "test-container"
	options.Config = config
	options.PodClient = clientset.CoreV1()
	options.Command = []string{"/bin/cp", "/dev/stdin", "/tmp/outfile"}

	err := options.Run()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (p *PodExec) PodCopyFile(src string, dst string, containername string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ioStreams, in, out, errOut := genericclioptions.NewTestIOStreams()
	copyOptions := cp.NewCopyOptions(ioStreams)
	copyOptions.Clientset = p.Clientset
	copyOptions.ClientConfig = p.RestConfig
	copyOptions.Container = containername
	err := copyOptions.Run([]string{src, dst})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not run copy operation: %v", err)
	}
	return in, out, errOut, nil
}
