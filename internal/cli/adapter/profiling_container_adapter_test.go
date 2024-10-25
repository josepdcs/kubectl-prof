package adapter

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	podexec "github.com/josepdcs/kubectl-prof/pkg/util/pod"
	"github.com/pkg/errors"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/handler"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes/job"
	resultfile "github.com/josepdcs/kubectl-prof/internal/cli/result"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func Test_profilingContainerAdapter_HandleProfilingContainerLogs(t *testing.T) {
	type fields struct {
		ProfilingContainerAdapter
	}
	type args struct {
		pod           *v1.Pod
		containerName string
		handler       EventHandler
		ctx           context.Context
	}
	type result struct {
		done       chan bool
		resultFile chan resultfile.File
		err        error
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) result
		then  func(t *testing.T, r result, f fields)
	}{
		{
			name: "should get result",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				return fields{
						NewProfilingContainerAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						handler:       handler.NewEventHandler(&config.TargetConfig{}, api.InfoLevel),
						ctx:           context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				done, resultFile, err := f.HandleProfilingContainerLogs(a.pod, a.containerName, a.handler, a.ctx)
				return result{
					done:       done,
					resultFile: resultFile,
					err:        err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.NoError(t, r.err)
			},
		},
		{
			name: "should fail when missing container name",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				return fields{
						NewProfilingContainerAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						pod:           pod,
						containerName: "",
						handler:       handler.NewEventHandler(&config.TargetConfig{}, api.InfoLevel),
						ctx:           context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				done, resultFile, err := f.HandleProfilingContainerLogs(a.pod, a.containerName, a.handler, a.ctx)
				return result{
					done:       done,
					resultFile: resultFile,
					err:        err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				assert.EqualError(t, r.err, "container name is mandatory for handling its logs")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			result := tt.when(fields, args)

			// Then
			tt.then(t, result, fields)
		})
	}
}

func Test_renameResultFileName(t *testing.T) {
	// Given
	podName := "pod-name"
	fileName := "/tmp/flamegraph.svg.gz"
	timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")

	// When
	result := renameResultFileName(podName, fileName, timestamp)

	// Then
	assert.Equal(t, "pod-name-flamegraph-2023-02-28T11_44_12Z.svg", result)
}

func Test_profilingContainerAdapter_GetRemoteFile(t *testing.T) {
	type fields struct {
		ProfilingContainerAdapter
	}
	type args struct {
		pod           *v1.Pod
		containerName string
		remoteFile    resultfile.File
		target        *config.TargetConfig
		localPath     string
		t             compressor.Type
	}
	type result struct {
		remoteFile string
		err        error
	}
	tests := []struct {
		name      string
		given     func() (fields, args)
		when      func(fields, args) result
		then      func(t *testing.T, r result, f fields)
		afterEach func()
	}{
		{
			name: "should get remote file",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				outFake := bytes.NewBufferString(remoteFileContent)
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(outFake, nil, nil),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte(remoteFileContent)),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
						},
						target: &config.TargetConfig{
							LocalPath:  "/tmp",
							Compressor: compressor.None,
							PodName:    "pod-name",
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.NoError(t, r.err)
				assert.Equal(t, filepath.Join(common.TmpDir(), "pod-name-flamegraph-2023-02-28T11_44_12Z.svg"), r.remoteFile)
			},
			afterEach: func() {
				_ = os.Remove(filepath.Join(common.TmpDir(), "pod-name-flamegraph-2023-02-28T11_44_12Z.svg"))
			},
		},
		{
			name: "should get remote fail when download failed",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				errOutFake := bytes.NewBufferString("error message")
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(nil, errOutFake, errors.New("error")),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte(remoteFileContent)),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
						},
						target: &config.TargetConfig{
							LocalPath:  "/tmp",
							Compressor: compressor.None,
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.EqualError(t, r.err, "could not download profiler result file from pod: error message: error")
			},
		},
		{
			name: "should get remote file fail when checksum is not equal",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				outFake := bytes.NewBufferString(remoteFileContent)
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(outFake, nil, nil),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte("other")),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
						},
						target: &config.TargetConfig{
							LocalPath:  "/tmp",
							Compressor: compressor.None,
							ExtraTargetOptions: config.ExtraTargetOptions{
								RetrieveFileRetries: 3,
							},
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.EqualError(t, r.err, "checksum does not match for file /tmp/flamegraph.svg.gz")
			},
		},
		{
			name: "should get remote file fail when compressor is unknown",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				outFake := bytes.NewBufferString(remoteFileContent)
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(outFake, nil, nil),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte(remoteFileContent)),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
						},
						target: &config.TargetConfig{
							LocalPath:  "/tmp",
							Compressor: "other",
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.EqualError(t, r.err, "could not get compressor: could not find compressor for other")
			},
		},
		{
			name: "should get remote file fail when uncompress failed",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				outFake := bytes.NewBufferString(remoteFileContent)
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(outFake, nil, nil),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte(remoteFileContent)),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
						},
						target: &config.TargetConfig{
							LocalPath:  "/tmp",
							Compressor: compressor.Gzip,
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.EqualError(t, r.err, "could not decode remote file: unexpected EOF")
			},
		},
		{
			name: "should get remote file fail when creation file failed",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				outFake := bytes.NewBufferString(remoteFileContent)
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(outFake, nil, nil),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte(remoteFileContent)),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
						},
						target: &config.TargetConfig{
							LocalPath:  "/other",
							Compressor: compressor.None,
							PodName:    "pod-name",
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.EqualError(t, r.err, "could not write result file: open /other/pod-name-flamegraph-2023-02-28T11_44_12Z.svg: no such file or directory")
			},
		},
		{
			name: "should get remote file when chunks are available",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				outFake := bytes.NewBufferString(remoteFileContent)
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(outFake, nil, nil),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte(remoteFileContent)),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
							Chunks: []api.ChunkData{
								{
									File:            filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz.00",
									FileSizeInBytes: int64(len(remoteFileContent)),
									Checksum:        getMD5Hash([]byte(remoteFileContent)),
								},
							},
						},
						target: &config.TargetConfig{
							LocalPath:  "/tmp",
							Compressor: compressor.None,
							PodName:    "pod-name",
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.NoError(t, r.err)
				assert.Equal(t, filepath.Join(common.TmpDir(), "pod-name-flamegraph-2023-02-28T11_44_12Z.svg"), r.remoteFile)
			},
			afterEach: func() {
				_ = os.Remove(filepath.Join(common.TmpDir(), "pod-name-flamegraph-2023-02-28T11_44_12Z.svg"))
			},
		},
		{
			name: "should get remote fail when download chunk failed",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				errOutFake := bytes.NewBufferString("error message")
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(nil, errOutFake, errors.New("error")),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte(remoteFileContent)),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
							Chunks: []api.ChunkData{
								{
									File:            filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz.00",
									FileSizeInBytes: int64(len(remoteFileContent)),
									Checksum:        getMD5Hash([]byte(remoteFileContent)),
								},
							},
						},
						target: &config.TargetConfig{
							LocalPath:  "/tmp",
							Compressor: compressor.None,
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.EqualError(t, r.err, "could not download profiler chunk file from pod: error message: error")
			},
		},
		{
			name: "should get remote file fail when checksum chunk is not equal",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				outFake := bytes.NewBufferString(remoteFileContent)
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(outFake, nil, nil),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte(remoteFileContent)),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
							Chunks: []api.ChunkData{
								{
									File:            filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz.00",
									FileSizeInBytes: int64(len(remoteFileContent)),
									Checksum:        getMD5Hash([]byte("other")),
								},
							},
						},
						target: &config.TargetConfig{
							LocalPath:  "/tmp",
							Compressor: compressor.None,
							ExtraTargetOptions: config.ExtraTargetOptions{
								RetrieveFileRetries: 3,
							},
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.EqualError(t, r.err, "checksum does not match for chunk file /tmp/flamegraph.svg.gz.00")
			},
		},
		{
			name: "should get remote file fail when chunks are not written to disk",
			given: func() (fields, args) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "PodName",
						GenerateName: "",
						Namespace:    "Namespace",
						Labels: map[string]string{
							job.LabelID: "Id",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "ContainerName",
								Image: "Image",
							},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				remoteFileContent := "test"
				outFake := bytes.NewBufferString(remoteFileContent)
				timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")
				return fields{
						profilingContainerAdapter{
							connectionInfo: kubernetes.ConnectionInfo{
								ClientSet: testclient.NewSimpleClientset(),
							},
							executor: podexec.NewExecFake(outFake, nil, nil),
						},
					},
					args{
						pod:           pod,
						containerName: "ContainerName",
						remoteFile: resultfile.File{
							FileName:        filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz",
							Checksum:        getMD5Hash([]byte(remoteFileContent)),
							Timestamp:       timestamp,
							FileSizeInBytes: int64(len(remoteFileContent)),
							Chunks: []api.ChunkData{
								{
									File:            filepath.Join(common.TmpDir(), "flamegraph.svg") + ".gz.00",
									FileSizeInBytes: int64(len(remoteFileContent)),
									Checksum:        getMD5Hash([]byte(remoteFileContent)),
								},
							},
						},
						target: &config.TargetConfig{
							LocalPath:  "/other",
							Compressor: compressor.None,
						},
					}
			},
			when: func(fields fields, args args) result {
				file, err := fields.GetRemoteFile(args.pod, args.containerName, args.remoteFile, args.target)
				return result{
					remoteFile: file,
					err:        err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.EqualError(t, r.err, "could not write chunk file: open /other/flamegraph-2023-02-28T11_44_12Z.svg.gz.00: no such file or directory")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			result := tt.when(fields, args)

			// Then
			tt.then(t, result, fields)

			if tt.afterEach != nil {
				tt.afterEach()
			}
		})
	}
}
