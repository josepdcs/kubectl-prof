package api

import (
	"context"
	"errors"
	"testing"

	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/rest"
	kubetesting "k8s.io/client-go/testing"
)

func Test_podAdapter_GetPod(t *testing.T) {
	type fields struct {
		PodApi
	}
	type args struct {
		podName   string
		namespace string
		ctx       context.Context
	}
	type result struct {
		pod *v1.Pod
		err error
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) result
		then  func(t *testing.T, r result, f fields)
	}{
		{
			name: "should get pod",
			given: func() (fields, args) {
				podList := &v1.PodList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:         "PodName",
								GenerateName: "",
								Namespace:    "Namespace",
								Labels: map[string]string{
									job.LabelID: "Id",
								},
							},
							Status: v1.PodStatus{
								Phase: v1.PodRunning,
							},
						},
					},
				}
				return fields{
						NewPodApi(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(podList),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						podName:   "PodName",
						namespace: "Namespace",
						ctx:       context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.GetPod(a.ctx, a.podName, a.namespace)
				return result{
					pod: gotPod,
					err: err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.NoError(t, r.err)
				assert.NotEmpty(t, r.pod)
			},
		},
		{
			name: "should not get pod",
			given: func() (fields, args) {
				clientSet := testclient.NewSimpleClientset()
				clientSet.CoreV1().(*fake.FakeCoreV1).PrependReactor("get", "pods", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("error getting pod")
				})
				return fields{
						PodApi: NewPodApi(
							kubernetes.ConnectionInfo{
								ClientSet:  clientSet,
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						podName:   "PodName",
						namespace: "Namespace",
						ctx:       context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.GetPod(a.ctx, a.podName, a.namespace)
				return result{
					pod: gotPod,
					err: err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Empty(t, &r.pod)
				assert.EqualError(t, r.err, "error getting pod")
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

func Test_podAdapter_GetPodsByLabelSelector(t *testing.T) {
	type fields struct {
		PodApi
	}
	type args struct {
		namespace      string
		labelSelectors string
		ctx            context.Context
	}
	type result struct {
		pods []v1.Pod
		err  error
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) result
		then  func(t *testing.T, r result, f fields)
	}{
		{
			name: "should get pods by label selector",
			given: func() (fields, args) {
				podList := &v1.PodList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:         "PodName",
								GenerateName: "",
								Namespace:    "Namespace",
								Labels: map[string]string{
									"selector": "app",
								},
							},
							Status: v1.PodStatus{
								Phase: v1.PodRunning,
							},
						},
					},
				}
				return fields{
						NewPodApi(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(podList),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						namespace:      "Namespace",
						labelSelectors: "selector=app",
						ctx:            context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPods, err := f.GetPodsByLabelSelector(a.ctx, a.namespace, a.labelSelectors)
				return result{
					pods: gotPods,
					err:  err,
				}
			},
			then: func(t *testing.T, r result, f fields) {
				require.NoError(t, r.err)
				assert.NotEmpty(t, r.pods)
			},
		},
		{
			name: "should not get pod",
			given: func() (fields, args) {
				clientSet := testclient.NewSimpleClientset()
				clientSet.CoreV1().(*fake.FakeCoreV1).PrependReactor("list", "pods", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("error getting pod")
				})
				return fields{
						PodApi: NewPodApi(
							kubernetes.ConnectionInfo{
								ClientSet:  clientSet,
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						namespace:      "Namespace",
						labelSelectors: "selector=app",
						ctx:            context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPods, err := f.GetPodsByLabelSelector(a.ctx, a.namespace, a.labelSelectors)
				return result{
					pods: gotPods,
					err:  err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Empty(t, &r.pods)
				assert.EqualError(t, r.err, "error getting pod")
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
