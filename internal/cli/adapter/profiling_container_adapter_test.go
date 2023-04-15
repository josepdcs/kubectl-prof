package adapter

import (
	"context"
	"testing"
	"time"

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
	fileName := "/tmp/contprof-flamegraph.svg.gz"
	timestamp, _ := time.Parse(time.RFC3339, "2023-02-28T11:44:12.678378359Z")

	// When
	result := renameResultFileName(fileName, timestamp)

	// Then
	assert.Equal(t, "contprof-flamegraph-2023-02-28T11_44_12Z.svg", result)
}
