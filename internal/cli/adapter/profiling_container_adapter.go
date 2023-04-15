package adapter

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/result"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	podexec "github.com/josepdcs/kubectl-prof/pkg/util/pod"
	v1 "k8s.io/api/core/v1"
)

type EventHandler interface {
	Handle(events chan string, done chan bool, resultFile chan result.File)
}

// ProfilingContainerAdapter defines all methods related to the profiling container
// A profiling container will be used for both a profiling job (and pod) and for ephemeral container
type ProfilingContainerAdapter interface {
	// HandleProfilingContainerLogs handles the logs of the profiling container up to obtain the result file if no error found
	HandleProfilingContainerLogs(pod *v1.Pod, containerName string, handler EventHandler, ctx context.Context) (chan bool, chan result.File, error)
	// GetRemoteFile returns the remote file from the pod's container
	GetRemoteFile(pod *v1.Pod, containerName string, remoteFile result.File, localPath string, c compressor.Type) (string, error)
}

// profilingContainerAdapter implements ProfilingContainerAdapter and wraps kubernetes.ConnectionInfo
type profilingContainerAdapter struct {
	connectionInfo kubernetes.ConnectionInfo
}

// NewProfilingContainerAdapter returns new instance of ProfilingContainerAdapter
func NewProfilingContainerAdapter(connectionInfo kubernetes.ConnectionInfo) ProfilingContainerAdapter {
	return profilingContainerAdapter{
		connectionInfo: connectionInfo,
	}
}

func (p profilingContainerAdapter) HandleProfilingContainerLogs(pod *v1.Pod, containerName string, handler EventHandler, ctx context.Context) (chan bool, chan result.File, error) {
	if stringUtils.IsBlank(containerName) {
		return nil, nil, errors.New("container name is mandatory for handling its logs")
	}
	req := p.connectionInfo.ClientSet.CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, &v1.PodLogOptions{
			Follow:    true,
			Container: containerName,
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

func (p profilingContainerAdapter) GetRemoteFile(pod *v1.Pod, containerName string, remoteFile result.File, localPath string, t compressor.Type) (string, error) {
	podFile := podexec.NewExec(p.connectionInfo.RestConfig, p.connectionInfo.ClientSet, pod, containerName)

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
	return stringUtils.SubstringBefore(f, ".") + "-" + strings.ReplaceAll(t.Format(time.RFC3339), ":", "_") + "." + stringUtils.SubstringAfter(f, ".")
}
