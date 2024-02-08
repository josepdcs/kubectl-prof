package adapter

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/alitto/pond"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
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
	GetRemoteFile(pod *v1.Pod, containerName string, remoteFile result.File, target *config.TargetConfig) (string, error)
}

// profilingContainerAdapter implements ProfilingContainerAdapter and wraps kubernetes.ConnectionInfo
type profilingContainerAdapter struct {
	connectionInfo kubernetes.ConnectionInfo
	executor       podexec.Executor
}

// NewProfilingContainerAdapter returns new instance of ProfilingContainerAdapter
func NewProfilingContainerAdapter(connectionInfo kubernetes.ConnectionInfo) ProfilingContainerAdapter {
	return profilingContainerAdapter{
		connectionInfo: connectionInfo,
		executor:       podexec.NewExec(connectionInfo.RestConfig, connectionInfo.ClientSet),
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

func (p profilingContainerAdapter) GetRemoteFile(pod *v1.Pod, containerName string, remoteFile result.File, target *config.TargetConfig) (string, error) {
	var fileBuff []byte

	if remoteFile.Chunks != nil && len(remoteFile.Chunks) > 0 {
		chunks, err := retrieveChunks(pod, containerName, remoteFile, p.executor, target)
		if err != nil {
			return "", err
		}

		fileBuff, err = readChunks(chunks, remoteFile.FileSizeInBytes)
		if err != nil {
			return "", err
		}
	} else {
		var err error
		fileBuff, err = retrieveFileOrRetry(pod, containerName, p.executor, remoteFile, target)
		if err != nil {
			return "", err
		}
	}

	comp, err := compressor.Get(target.Compressor)
	if err != nil {
		return "", errors.Wrap(err, "could not get compressor")
	}

	decoded, err := comp.Decode(fileBuff)
	if err != nil {
		return "", errors.Wrap(err, "could not decode remote file")
	}

	fileName := filepath.Join(target.LocalPath, renameResultFileName(remoteFile.FileName, remoteFile.Timestamp))

	err = os.WriteFile(fileName, decoded, 0644)
	if err != nil {
		return "", errors.Wrap(err, "could not write result file")
	}

	return fileName, nil
}

// retrieveChunks retrieves the chunks of the remote file from the pod's container
func retrieveChunks(pod *v1.Pod, containerName string, remoteFile result.File, exec podexec.Executor, target *config.TargetConfig) ([]string, error) {
	downloadChunks := make([]string, 0, len(remoteFile.Chunks))

	pool := pond.New(target.PoolSizeRetrieveChunks, 0, pond.MinWorkers(target.PoolSizeRetrieveChunks))
	defer pool.StopAndWait()

	// create a task group associated to a context
	group, _ := pool.GroupContext(context.Background())

	// submit tasks to download chunks
	for _, chunk := range remoteFile.Chunks {
		chunk := chunk
		group.Submit(func() error {
			fileName, err := retrieveChunkOrRetry(chunk, pod, containerName, exec, target, remoteFile.Timestamp)
			if err == nil {
				downloadChunks = append(downloadChunks, fileName)
			}
			return err
		})
	}

	// wait for all tasks to finish
	err := group.Wait()

	return downloadChunks, err
}

func retrieveFileOrRetry(pod *v1.Pod, containerName string, exec podexec.Executor, remoteFile result.File, target *config.TargetConfig) ([]byte, error) {
	fileBuff := make([]byte, 0, remoteFile.FileSizeInBytes)
	for i := 0; i <= target.RetrieveFileRetries; i++ {
		offset := 0
		n := 0
		for (remoteFile.FileSizeInBytes - int64(n)) > 0 {
			_, out, errOut, err := exec.Execute(pod.Namespace, pod.Name, containerName, []string{"sh", "-c", fmt.Sprintf("tail -c+%d %s", offset, remoteFile.FileName)})
			if err != nil {
				return nil, errors.Wrapf(err, "could not download profiler result file from pod: %s", errOut.String())
			}
			n += out.Len()
			offset += out.Len() + 1
			fileBuff = append(fileBuff, out.Bytes()...)
		}

		// check the checksum of the downloaded file
		checksum := getMD5Hash(fileBuff)
		log.Debugf("File %s downloaded (local: %s (%d bytes) | remote: %s (%d bytes))", remoteFile.FileName, checksum, len(fileBuff), remoteFile.Checksum, remoteFile.FileSizeInBytes)

		if checksum == remoteFile.Checksum {
			return fileBuff, nil
		}
		log.Debugf("Checksum does not match, retrying: %s...", remoteFile.FileName)

		fileBuff = make([]byte, 0, remoteFile.FileSizeInBytes)
	}

	// if the checksum does not match after the last retry, return an error
	return nil, errors.Errorf("checksum does not match for file %s", remoteFile.FileName)
}

func retrieveChunkOrRetry(chunk api.ChunkData, pod *v1.Pod, containerName string, exec podexec.Executor, target *config.TargetConfig, timestamp time.Time) (string, error) {
	for i := 0; i <= target.RetrieveFileRetries; i++ {
		// download the chunk file
		fileBuff, err := retrieveChunk(chunk, pod, containerName, exec)
		if err != nil {
			return "", err
		}

		// check the checksum of the downloaded chunk
		checksum := getMD5Hash(fileBuff)
		log.Debugf("Chunk file %s downloaded (local: %s (%d bytes) | remote: %s (%d bytes))", chunk.File, checksum, len(fileBuff), chunk.Checksum, chunk.FileSizeInBytes)

		// if the checksum matches, write the chunk file to the local filesystem
		// and return the file name
		if checksum == chunk.Checksum {
			fileName := filepath.Join(target.LocalPath, renameChunkFileName(chunk.File, timestamp))
			err = os.WriteFile(fileName, fileBuff, 0644)
			if err != nil {
				return "", errors.Wrap(err, "could not write chunk file")
			}

			return fileName, nil
		}
		log.Debugf("Checksum does not match, retrying: %s...", chunk.File)
	}

	// if the checksum does not match after the last retry, return an error
	return "", errors.Errorf("checksum does not match for chunk file %s", chunk.File)
}

// retrieveChunk retrieves the chunk of the remote file from the pod's container
func retrieveChunk(chunk api.ChunkData, pod *v1.Pod, containerName string, exec podexec.Executor) ([]byte, error) {
	fileBuff := make([]byte, 0, chunk.FileSizeInBytes)
	offset := 0
	n := 0
	log.Debugf("Downloading chunk file %s ...", chunk.File)
	for (chunk.FileSizeInBytes - int64(n)) > 0 {
		_, out, errOut, err := exec.Execute(pod.Namespace, pod.Name, containerName, []string{"sh", "-c", fmt.Sprintf("tail -c+%d %s", offset, chunk.File)})
		if err != nil {
			return nil, errors.Wrapf(err, "could not download profiler chunk file from pod: %s", errOut.String())
		}
		n += out.Len()
		offset += out.Len() + 1
		fileBuff = append(fileBuff, out.Bytes()...)
	}
	return fileBuff, nil
}

// readChunks reads the download chunks of the remote file from the pod's container.
// Each chunk is appended to the fileBuff which is returned
func readChunks(downloadChunks []string, fileBuffSize int64) ([]byte, error) {
	fileBuff := make([]byte, 0, fileBuffSize)
	// be sure that the chunks are sorted
	slices.Sort(downloadChunks)
	for _, downloadChunk := range downloadChunks {
		chunkFile, err := os.Open(downloadChunk)
		if err != nil {
			return nil, errors.Wrap(err, "could not open chunk file")
		}

		reader := bufio.NewReader(chunkFile)
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, errors.Wrap(err, "could not read chunk file")
		}

		fileBuff = append(fileBuff, content...)

		_ = chunkFile.Close()

		err = os.Remove(downloadChunk)
		if err != nil {
			return nil, errors.Wrapf(err, "could not remove chunk file")
		}
	}
	return fileBuff, nil
}

// renameResultFileName renames the result file
func renameResultFileName(fileName string, t time.Time) string {
	f := stringUtils.SubstringBeforeLast(stringUtils.SubstringAfterLast(fileName, "/"), ".")
	return stringUtils.SubstringBefore(f, ".") + "-" + strings.ReplaceAll(t.Format(time.RFC3339), ":", "_") + "." + stringUtils.SubstringAfter(f, ".")
}

// renameChunkFileName renames the chunk file
func renameChunkFileName(fileName string, t time.Time) string {
	ending := "." + stringUtils.SubstringAfterLast(fileName, ".")
	f := stringUtils.RemoveEnd(stringUtils.SubstringAfterLast(fileName, "/"), ending)
	return stringUtils.SubstringBefore(f, ".") + "-" + strings.ReplaceAll(t.Format(time.RFC3339), ":", "_") + "." + stringUtils.SubstringAfter(f, ".") + ending
}

// getMD5Hash returns the MD5 hash of the given text
func getMD5Hash(text []byte) string {
	hash := md5.Sum(text)
	return hex.EncodeToString(hash[:])
}
