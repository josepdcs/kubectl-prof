package adapter

import (
	"context"
	"errors"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
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
	"testing"
	"time"
)

func Test_profilingEphemeralContainerAdapter_AddEphemeralContainer(t *testing.T) {
	type fields struct {
		ProfilingEphemeralContainerAdapter
	}
	type args struct {
		targetPod *v1.Pod
		cfg       *config.ProfilerConfig
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
			name: "should add ephemeral container",
			given: func() (fields, args) {
				podList := &v1.PodList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PodName",
								Namespace: "Namespace",
								Labels: map[string]string{
									job.LabelID: "Id",
								},
							},
							Spec: v1.PodSpec{
								EphemeralContainers: nil,
							},
							Status: v1.PodStatus{
								Phase: v1.PodRunning,
								EphemeralContainerStatuses: []v1.ContainerStatus{
									{
										Name: "EphemeralContainerName",
										State: v1.ContainerState{
											Running: &v1.ContainerStateRunning{},
										},
									},
								},
							},
						},
					},
				}
				return fields{
						ProfilingEphemeralContainerAdapter: NewProfilingEphemeralContainerAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(podList),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						targetPod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PodName",
								Namespace: "Namespace",
								Labels: map[string]string{
									job.LabelID: "Id",
								},
							},
							Status: v1.PodStatus{
								Phase: v1.PodRunning,
							},
						},
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								DryRun:   false,
								Language: api.FakeLang,
							},
							EphemeralContainer: &config.EphemeralContainerConfig{Privileged: true},
						},
						ctx: context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.AddEphemeralContainer(a.targetPod, a.cfg, a.ctx, 1*time.Second)
				return result{
					pod: gotPod,
					err: err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.NoError(t, r.err)
				assert.NotEmpty(t, r.pod)
				assert.Equal(t, "PodName", r.pod.GetName())
				assert.Equal(t, "Namespace", r.pod.GetNamespace())
				assert.NotEmpty(t, r.pod.Spec.EphemeralContainers)
			},
		},
		{
			name: "should print pod with added ephemeral container",
			given: func() (fields, args) {
				return fields{
						ProfilingEphemeralContainerAdapter: NewProfilingEphemeralContainerAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						targetPod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PodName",
								Namespace: "Namespace",
								Labels: map[string]string{
									job.LabelID: "Id",
								},
							},
							Status: v1.PodStatus{
								Phase: v1.PodRunning,
							},
						},
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								DryRun:   true,
								Language: api.FakeLang,
							},
							EphemeralContainer: &config.EphemeralContainerConfig{Privileged: true},
						},
						ctx: context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.AddEphemeralContainer(a.targetPod, a.cfg, a.ctx, 1*time.Second)
				return result{
					pod: gotPod,
					err: err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.NoError(t, r.err)
				assert.Empty(t, r.pod)
			},
		},
		{
			name: "should fail when unable getting ephemeral container type",
			given: func() (fields, args) {
				return fields{
						ProfilingEphemeralContainerAdapter: NewProfilingEphemeralContainerAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						targetPod: &v1.Pod{},
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								DryRun:   false,
								Language: "unexpected",
							},
							Job: &config.JobConfig{
								Namespace: "Namespace",
							},
						},
						ctx: context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.AddEphemeralContainer(a.targetPod, a.cfg, a.ctx, 1*time.Second)
				return result{
					pod: gotPod,
					err: err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Empty(t, r.pod)
				assert.EqualError(t, r.err, "unable to get the ephemeral container type: got language without job creator")
			},
		},
		{
			name: "should fail when fail patch for not found reason",
			given: func() (fields, args) {
				return fields{
						ProfilingEphemeralContainerAdapter: NewProfilingEphemeralContainerAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						targetPod: &v1.Pod{},
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								DryRun:   false,
								Language: api.FakeLang,
							},
							Job: &config.JobConfig{
								Namespace: "Namespace",
							},
						},
						ctx: context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.AddEphemeralContainer(a.targetPod, a.cfg, a.ctx, 1*time.Second)
				return result{
					pod: gotPod,
					err: err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Empty(t, r.pod)
				assert.ErrorContains(t, r.err, "ephemeral containers are disabled for this cluster")
			},
		},
		{
			name: "should fail when fail patch for unexpected reason",
			given: func() (fields, args) {
				clientSet := testclient.NewSimpleClientset()
				clientSet.CoreV1().(*fake.FakeCoreV1).PrependReactor("patch", "pods", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.Pod{}, errors.New("error patching pod")
				})
				return fields{
						ProfilingEphemeralContainerAdapter: NewProfilingEphemeralContainerAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  clientSet,
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						targetPod: &v1.Pod{},
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								DryRun:   false,
								Language: api.FakeLang,
							},
							Job: &config.JobConfig{
								Namespace: "Namespace",
							},
						},
						ctx: context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.AddEphemeralContainer(a.targetPod, a.cfg, a.ctx, 1*time.Second)
				return result{
					pod: gotPod,
					err: err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Empty(t, r.pod)
				assert.EqualError(t, r.err, "error patching pod")
			},
		},
		{
			name: "should fail when ephemeral container name not match",
			given: func() (fields, args) {
				podList := &v1.PodList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PodName",
								Namespace: "Namespace",
								Labels: map[string]string{
									job.LabelID: "Id",
								},
							},
							Spec: v1.PodSpec{
								EphemeralContainers: nil,
							},
							Status: v1.PodStatus{
								Phase: v1.PodRunning,
								EphemeralContainerStatuses: []v1.ContainerStatus{
									{
										Name: "OtherEphemeralContainerName",
										State: v1.ContainerState{
											Running: &v1.ContainerStateRunning{},
										},
									},
								},
							},
						},
					},
				}
				return fields{
						ProfilingEphemeralContainerAdapter: NewProfilingEphemeralContainerAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(podList),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						targetPod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PodName",
								Namespace: "Namespace",
								Labels: map[string]string{
									job.LabelID: "Id",
								},
							},
							Status: v1.PodStatus{
								Phase: v1.PodRunning,
							},
						},
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								DryRun:   false,
								Language: api.FakeLang,
							},
							EphemeralContainer: &config.EphemeralContainerConfig{Privileged: true},
						},
						ctx: context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.AddEphemeralContainer(a.targetPod, a.cfg, a.ctx, 1*time.Second)
				return result{
					pod: gotPod,
					err: err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Empty(t, r.pod)
				assert.EqualError(t, r.err, "timed out waiting for the condition")
			},
		},
		{
			name: "should fail when patched pod failed",
			given: func() (fields, args) {
				podList := &v1.PodList{
					TypeMeta: metav1.TypeMeta{},
					ListMeta: metav1.ListMeta{},
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PodName",
								Namespace: "Namespace",
								Labels: map[string]string{
									job.LabelID: "Id",
								},
							},
							Spec: v1.PodSpec{
								EphemeralContainers: nil,
							},
							Status: v1.PodStatus{
								Phase: v1.PodFailed,
								EphemeralContainerStatuses: []v1.ContainerStatus{
									{
										Name: "EphemeralContainerName",
										State: v1.ContainerState{
											Running: &v1.ContainerStateRunning{},
										},
									},
								},
							},
						},
					},
				}
				return fields{
						ProfilingEphemeralContainerAdapter: NewProfilingEphemeralContainerAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(podList),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						targetPod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PodName",
								Namespace: "Namespace",
								Labels: map[string]string{
									job.LabelID: "Id",
								},
							},
							Status: v1.PodStatus{
								Phase: v1.PodRunning,
							},
						},
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								DryRun:   false,
								Language: api.FakeLang,
							},
							EphemeralContainer: &config.EphemeralContainerConfig{Privileged: true},
						},
						ctx: context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.AddEphemeralContainer(a.targetPod, a.cfg, a.ctx, 1*time.Second)
				return result{
					pod: gotPod,
					err: err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Empty(t, r.pod)
				assert.EqualError(t, r.err, "target pod now is failed: PodName")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			r := tt.when(fields, args)

			// Then
			tt.then(t, r, fields)
		})
	}

}
