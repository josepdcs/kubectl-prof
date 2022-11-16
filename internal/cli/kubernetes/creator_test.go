package kubernetes

import (
	"context"
	"github.com/josepdcs/kubectl-prof/api"
	config2 "github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestNewCreator(t *testing.T) {
	assert.IsType(t, &JobCreator{}, NewJobCreator(testclient.NewSimpleClientset()))
}

type mockCreator struct {
	mock.Mock
}

func (m *mockCreator) Create(targetPod *v1.Pod, cfg *config2.ProfilerConfig) (string, *batchv1.Job, error) {
	args := m.Called(targetPod, cfg)
	return args.String(0), args.Get(1).(*batchv1.Job), args.Error(2)
}

func Test_creator_CreateProfilingJob(t *testing.T) {
	type args struct {
		targetPod *v1.Pod
		cfg       *config2.ProfilerConfig
		ctx       context.Context
	}
	tests := []struct {
		name        string
		args        args
		mockJobType func()
		wantId      string
		wantJob     *batchv1.Job
		wantErrMsg  string
	}{
		{
			name: "should create profiling job",
			args: args{
				targetPod: &v1.Pod{},
				cfg: &config2.ProfilerConfig{
					Target: &config2.TargetConfig{
						DryRun:   false,
						Language: api.Java,
					},
					Job: &config2.JobConfig{
						Namespace: "Namespace",
					},
				},
				ctx: context.TODO(),
			},
			mockJobType: func() {
				jobType = func(language api.ProgrammingLanguage, tool api.ProfilingTool) (job.Creator, error) {
					m := mockCreator{}
					m.On("Create", mock.Anything, mock.Anything).Return("ID", &batchv1.Job{}, nil)
					return &m, nil
				}
			},
			wantId: "ID",
			wantJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "Namespace",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockJobType()

			c := JobCreator{
				clientSet: testclient.NewSimpleClientset(),
			}
			gotId, gotJob, err := c.CreateProfilingJob(tt.args.targetPod, tt.args.cfg, tt.args.ctx)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantId, gotId)
			assert.Equal(t, tt.wantJob, gotJob)
		})
	}
}
