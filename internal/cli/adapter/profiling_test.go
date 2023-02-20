package adapter

import (
	"context"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"testing"
)

func Test_profilingAdapter_CreateProfilingJob(t *testing.T) {
	type fields struct {
		ProfilingAdapter
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

		fields      fields
		args        args
		mockJobType func()
		wantId      string
		wantJob     *batchv1.Job
		wantErrMsg  string
	}{
		{
			name: "should create profiling job",
			given: func() (fields, args) {
				return fields{
						NewProfilingAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						targetPod: &v1.Pod{},
						cfg: config.NewProfilerConfig(
							&config.TargetConfig{
								DryRun:   false,
								Language: api.FakeLang,
							},
							&config.JobConfig{
								Namespace: "Namespace",
							},
						),
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
						NewProfilingAdapter(
							kubernetes.ConnectionInfo{
								ClientSet:  testclient.NewSimpleClientset(),
								RestConfig: &rest.Config{},
								Namespace:  "Namespace",
							},
						),
					},
					args{
						targetPod: &v1.Pod{},
						cfg: config.NewProfilerConfig(
							&config.TargetConfig{
								DryRun:   false,
								Language: "unexpected",
							},
							&config.JobConfig{
								Namespace: "Namespace",
							},
						),
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
				assert.EqualError(t, r.err, "unable to get type of job: got language without job creator")
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
