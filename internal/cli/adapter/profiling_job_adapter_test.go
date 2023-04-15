package adapter

import (
	"context"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"testing"
	"time"
)

func Test_profilingJobAdapter_CreateProfilingJob(t *testing.T) {
	type fields struct {
		ProfilingJobAdapter
	}
	type args struct {
		targetPod *v1.Pod
		cfg       *config.ProfilerConfig
		ctx       context.Context
	}
	type result struct {
		jobId        string
		profilingJob *batchv1.Job
		err          error
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) result
		then  func(t *testing.T, r result, f fields)
	}{
		{
			name: "should create profiling job",
			given: func() (fields, args) {
				return fields{
						NewProfilingJobAdapter(
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
				gotId, gotJob, err := f.CreateProfilingJob(a.targetPod, a.cfg, a.ctx)
				return result{
					jobId:        gotId,
					profilingJob: gotJob,
					err:          err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				expected := result{
					jobId: "ID",
					profilingJob: &batchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "Namespace",
						},
					},
					err: nil,
				}
				assert.Equal(t, expected, r)
			},
		},
		{
			name: "should fail when unable getting job type",
			given: func() (fields, args) {
				return fields{
						NewProfilingJobAdapter(
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
				gotId, gotJob, err := f.CreateProfilingJob(a.targetPod, a.cfg, a.ctx)
				return result{
					jobId:        gotId,
					profilingJob: gotJob,
					err:          err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Equal(t, "", r.jobId)
				assert.Empty(t, r.profilingJob)
				assert.EqualError(t, r.err, "unable to get the job type: got language without job creator")
			},
		},
		{
			name: "should fail when build create job fail",
			given: func() (fields, args) {
				return fields{
						NewProfilingJobAdapter(
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
								Name: "PodError",
							},
						},
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								DryRun:        false,
								Language:      api.FakeLang,
								ProfilingTool: api.FakeTool,
							},
							Job: &config.JobConfig{
								Namespace: "Namespace",
							},
						},
						ctx: context.TODO(),
					}
			},
			when: func(f fields, a args) result {
				gotId, gotJob, err := f.CreateProfilingJob(a.targetPod, a.cfg, a.ctx)
				return result{
					jobId:        gotId,
					profilingJob: gotJob,
					err:          err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Equal(t, "", r.jobId)
				assert.Empty(t, r.profilingJob)
				assert.EqualError(t, r.err, "unable to create job: unable create job")
			},
		},
		{
			name: "should print profiling job",
			given: func() (fields, args) {
				return fields{
						NewProfilingJobAdapter(
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
								DryRun:   true,
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
				gotId, gotJob, err := f.CreateProfilingJob(a.targetPod, a.cfg, a.ctx)
				return result{
					jobId:        gotId,
					profilingJob: gotJob,
					err:          err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				expected := result{
					jobId:        "",
					profilingJob: nil,
					err:          nil,
				}
				assert.Equal(t, expected, r)
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

func Test_profilingJobAdapter_GetProfilingPod(t *testing.T) {
	type fields struct {
		ProfilingJobAdapter
	}
	type args struct {
		cfg     *config.ProfilerConfig
		ctx     context.Context
		timeout time.Duration
	}
	type result struct {
		profilingPod *v1.Pod
		err          error
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) result
		then  func(t *testing.T, r result, f fields)
	}{
		{
			name: "should get profiling pod",
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
						NewProfilingJobAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(podList),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Id:       "Id",
								DryRun:   false,
								Language: api.FakeLang,
							},
							Job: &config.JobConfig{
								Namespace: "Namespace",
							},
						},
						ctx:     context.TODO(),
						timeout: 1 * time.Second,
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.GetProfilingPod(a.cfg, a.ctx, a.timeout)
				return result{
					profilingPod: gotPod,
					err:          err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.NoError(t, r.err)
				assert.NotEmpty(t, r.profilingPod)
			},
		},
		{
			name: "should fail when profiling pod failed",
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
								Phase: v1.PodFailed,
							},
						},
					},
				}
				return fields{
						NewProfilingJobAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(podList),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								Id:       "Id",
								DryRun:   false,
								Language: api.FakeLang,
							},
							Job: &config.JobConfig{
								Namespace: "Namespace",
							},
						},
						ctx:     context.TODO(),
						timeout: 1 * time.Second,
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.GetProfilingPod(a.cfg, a.ctx, a.timeout)
				return result{
					profilingPod: gotPod,
					err:          err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Empty(t, r.profilingPod)
				assert.EqualError(t, r.err, "profiling pod failed")
			},
		},
		{
			name: "should fail for timeout",
			given: func() (fields, args) {
				return fields{
						NewProfilingJobAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						cfg: &config.ProfilerConfig{
							Target: &config.TargetConfig{
								DryRun:   false,
								Language: api.FakeLang,
							},
							Job: &config.JobConfig{
								Namespace: "Namespace",
							},
						},
						ctx:     context.TODO(),
						timeout: 1 * time.Second,
					}
			},
			when: func(f fields, a args) result {
				gotPod, err := f.GetProfilingPod(a.cfg, a.ctx, a.timeout)
				return result{
					profilingPod: gotPod,
					err:          err,
				}

			},
			then: func(t *testing.T, r result, f fields) {
				require.Error(t, r.err)
				assert.Empty(t, r.profilingPod)
				assert.EqualError(t, r.err, "timed out waiting for the condition")
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

func Test_profilingJobAdapter_GetProfilingContainerName(t *testing.T) {
	// Given & When
	result := NewProfilingJobAdapter(
		kubernetes.ConnectionInfo{
			ClientSet:  testclient.NewSimpleClientset(),
			RestConfig: &rest.Config{},
			Namespace:  "Namespace",
		},
	).GetProfilingContainerName()

	// Then
	assert.Equal(t, job.ContainerName, result)
}

func Test_profilingJobAdapter_DeleteProfilingJob(t *testing.T) {
	// Given & When
	j := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Name",
			Namespace: "Namespace",
		},
	}
	result := NewProfilingJobAdapter(
		kubernetes.ConnectionInfo{
			ClientSet:  testclient.NewSimpleClientset(j),
			RestConfig: &rest.Config{},
			Namespace:  "Namespace",
		},
	).DeleteProfilingJob(j, context.TODO())

	require.NoError(t, result)
}
