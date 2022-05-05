package pod

import (
	"bytes"
	"io"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/exec"
)

// PodFile
// Implement Read and Write interface
type PodFile struct {
	Path string
	*PodExec
}

func NewPodFile(path string, podexec *PodExec) *PodFile {
	return &PodFile{
		Path:    path,
		PodExec: podexec,
	}
}

// Write p []byte to Path
func (pf *PodFile) Write(b []byte) (n int, err error) {
	err = pf.uploadFile(bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

// Read from Path to b []byte
func (pf *PodFile) Read(b []byte) (n int, err error) {
	buf := bytes.NewBuffer([]byte{})
	written, err := pf.downloadFile(buf)
	if err != nil {
		return 0, err
	}
	copy(b, buf.Bytes())
	return int(written), io.EOF
}

func (pf *PodFile) uploadFile(r io.Reader) error {
	options := &exec.ExecOptions{}
	out := bytes.NewBuffer([]byte{})
	errOut := bytes.NewBuffer([]byte{})
	reader, writer := io.Pipe()

	go func(r io.Reader, writer *io.PipeWriter) {
		defer writer.Close()
		_, _ = io.Copy(writer, r)
	}(r, writer)

	options.StreamOptions = exec.StreamOptions{
		IOStreams: genericclioptions.IOStreams{
			In:     reader,
			Out:    out,
			ErrOut: errOut,
		},
		Stdin:     true,
		Namespace: pf.Namespace,
		PodName:   pf.PodName,
	}
	options.Executor = &exec.DefaultRemoteExecutor{}
	options.Namespace = pf.Namespace
	options.PodName = pf.PodName
	options.ContainerName = pf.ContainerName
	options.Config = pf.RestConfig
	options.PodClient = pf.Clientset.CoreV1()
	options.Command = []string{"tee", "-a", pf.Path}

	err := options.Run()
	if err != nil {
		return err
	}
	return nil
}

func (pf *PodFile) downloadFile(w io.Writer) (int64, error) {
	options := &exec.ExecOptions{}
	errOut := bytes.NewBuffer([]byte{})
	reader, writer := io.Pipe()

	options.StreamOptions = exec.StreamOptions{
		IOStreams: genericclioptions.IOStreams{
			In:     nil,
			Out:    writer,
			ErrOut: errOut,
		},
		Namespace: pf.Namespace,
		PodName:   pf.PodName,
	}
	options.Executor = &exec.DefaultRemoteExecutor{}
	options.Namespace = pf.Namespace
	options.PodName = pf.PodName
	options.ContainerName = pf.ContainerName
	options.Config = pf.PodExec.RestConfig
	options.PodClient = pf.PodExec.Clientset.CoreV1()
	options.Command = []string{"/bin/cp", pf.Path, "/dev/stdout"}

	go func(options *exec.ExecOptions, writer *io.PipeWriter) {
		defer writer.Close()
		_ = options.Run()
	}(options, writer)

	return io.Copy(w, reader)
}
