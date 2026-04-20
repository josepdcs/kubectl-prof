package profiler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	resp *http.Response
	err  error
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	return m.resp, m.err
}

func TestPprofProfiler_SetUp(t *testing.T) {
	tests := []struct {
		name    string
		job     *job.ProfilingJob
		wantErr bool
		errMsg  string
	}{
		{
			name: "should setup with pprof host",
			job: &job.ProfilingJob{
				AdditionalArguments: map[string]string{
					PprofHostKey: "10.0.0.1",
				},
			},
			wantErr: false,
		},
		{
			name: "should setup with pprof host and port",
			job: &job.ProfilingJob{
				AdditionalArguments: map[string]string{
					PprofHostKey: "10.0.0.1",
					PprofPortKey: "8080",
				},
			},
			wantErr: false,
		},
		{
			name: "should fail without pprof host",
			job: &job.ProfilingJob{
				AdditionalArguments: map[string]string{},
			},
			wantErr: true,
			errMsg:  "pprof host is required",
		},
		{
			name: "should fail with nil additional arguments",
			job: &job.ProfilingJob{
				AdditionalArguments: nil,
			},
			wantErr: true,
			errMsg:  "pprof host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profiler := NewPprofProfiler(nil)
			err := profiler.SetUp(tt.job)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPprofProfiler_Invoke(t *testing.T) {
	type fields struct {
		PprofProfiler *PprofProfiler
	}
	type args struct {
		job *job.ProfilingJob
	}
	tests := []struct {
		name  string
		given func() (fields, args)
		when  func(fields, args) (error, time.Duration)
		then  func(t *testing.T, err error, d time.Duration)
	}{
		{
			name: "should invoke with mock manager",
			given: func() (fields, args) {
				mgr := newMockPprofManager()
				mgr.On("invoke", &job.ProfilingJob{
					Tool:       api.GoPprof,
					OutputType: api.Raw,
					Interval:   30 * time.Second,
					Compressor: compressor.Gzip,
					Iteration:  1,
					AdditionalArguments: map[string]string{
						PprofHostKey: "10.0.0.1",
					},
				}).Return(nil, 1*time.Second)
				return fields{
						PprofProfiler: &PprofProfiler{PprofManager: mgr},
					}, args{
						job: &job.ProfilingJob{
							Tool:       api.GoPprof,
							OutputType: api.Raw,
							Interval:   30 * time.Second,
							Compressor: compressor.Gzip,
							Iteration:  1,
							AdditionalArguments: map[string]string{
								PprofHostKey: "10.0.0.1",
							},
						},
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PprofProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, d time.Duration) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should return error when invoke fails",
			given: func() (fields, args) {
				mgr := newMockPprofManager()
				j := &job.ProfilingJob{
					Tool:       api.GoPprof,
					OutputType: api.Raw,
					Interval:   30 * time.Second,
					Compressor: compressor.Gzip,
					Iteration:  1,
					AdditionalArguments: map[string]string{
						PprofHostKey: "10.0.0.1",
					},
				}
				mgr.On("invoke", j).Return(errors.New("connection refused"), 0*time.Second)
				return fields{
						PprofProfiler: &PprofProfiler{PprofManager: mgr},
					}, args{
						job: j,
					}
			},
			when: func(fields fields, args args) (error, time.Duration) {
				return fields.PprofProfiler.Invoke(args.job)
			},
			then: func(t *testing.T, err error, d time.Duration) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			fields, args := tt.given()

			// When
			err, d := tt.when(fields, args)

			// Then
			tt.then(t, err, d)
		})
	}
}

func TestPprofManager_invoke(t *testing.T) {
	tests := []struct {
		name       string
		httpClient *mockHTTPClient
		job        *job.ProfilingJob
		wantErr    bool
		errMsg     string
	}{
		{
			name: "should invoke successfully",
			httpClient: &mockHTTPClient{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte("profiling data here"))),
				},
			},
			job: &job.ProfilingJob{
				Tool:       api.GoPprof,
				OutputType: api.Raw,
				Interval:   30 * time.Second,
				Compressor: compressor.None,
				Iteration:  1,
				AdditionalArguments: map[string]string{
					PprofHostKey: "10.0.0.1",
					PprofPortKey: "6060",
				},
			},
			wantErr: false,
		},
		{
			name: "should invoke with default port",
			httpClient: &mockHTTPClient{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte("profiling data here"))),
				},
			},
			job: &job.ProfilingJob{
				Tool:       api.GoPprof,
				OutputType: api.Raw,
				Interval:   30 * time.Second,
				Compressor: compressor.None,
				Iteration:  1,
				AdditionalArguments: map[string]string{
					PprofHostKey: "10.0.0.1",
				},
			},
			wantErr: false,
		},
		{
			name: "should fail on HTTP error",
			httpClient: &mockHTTPClient{
				err: errors.New("connection refused"),
			},
			job: &job.ProfilingJob{
				Tool:       api.GoPprof,
				OutputType: api.Raw,
				Interval:   30 * time.Second,
				Compressor: compressor.None,
				Iteration:  1,
				AdditionalArguments: map[string]string{
					PprofHostKey: "10.0.0.1",
				},
			},
			wantErr: true,
			errMsg:  "failed to retrieve pprof data",
		},
		{
			name: "should fail on non-200 status",
			httpClient: &mockHTTPClient{
				resp: &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader([]byte(""))),
				},
			},
			job: &job.ProfilingJob{
				Tool:       api.GoPprof,
				OutputType: api.Raw,
				Interval:   30 * time.Second,
				Compressor: compressor.None,
				Iteration:  1,
				AdditionalArguments: map[string]string{
					PprofHostKey: "10.0.0.1",
				},
			},
			wantErr: true,
			errMsg:  "pprof endpoint returned HTTP 404",
		},
		{
			name: "should fail on empty response",
			httpClient: &mockHTTPClient{
				resp: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(""))),
				},
			},
			job: &job.ProfilingJob{
				Tool:       api.GoPprof,
				OutputType: api.Raw,
				Interval:   30 * time.Second,
				Compressor: compressor.None,
				Iteration:  1,
				AdditionalArguments: map[string]string{
					PprofHostKey: "10.0.0.1",
				},
			},
			wantErr: true,
			errMsg:  "pprof endpoint returned empty response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &pprofManager{
				publisher:  &fakePublisher{},
				httpClient: tt.httpClient,
			}

			err, _ := mgr.invoke(tt.job)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPprofProfiler_CleanUp(t *testing.T) {
	p := NewPprofProfiler(nil)
	err := p.CleanUp(&job.ProfilingJob{})
	assert.NoError(t, err)
}

func TestGetPprofHost(t *testing.T) {
	tests := []struct {
		name     string
		job      *job.ProfilingJob
		expected string
	}{
		{
			name: "should return host",
			job: &job.ProfilingJob{
				AdditionalArguments: map[string]string{
					PprofHostKey: "10.0.0.1",
				},
			},
			expected: "10.0.0.1",
		},
		{
			name:     "should return empty for nil args",
			job:      &job.ProfilingJob{},
			expected: "",
		},
		{
			name: "should return empty when key not present",
			job: &job.ProfilingJob{
				AdditionalArguments: map[string]string{},
			},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getPprofHost(tt.job))
		})
	}
}

func TestGetPprofPort(t *testing.T) {
	tests := []struct {
		name     string
		job      *job.ProfilingJob
		expected string
	}{
		{
			name: "should return custom port",
			job: &job.ProfilingJob{
				AdditionalArguments: map[string]string{
					PprofPortKey: "8080",
				},
			},
			expected: "8080",
		},
		{
			name: "should return default port when not set",
			job: &job.ProfilingJob{
				AdditionalArguments: map[string]string{},
			},
			expected: defaultPprofPort,
		},
		{
			name:     "should return default port for nil args",
			job:      &job.ProfilingJob{},
			expected: defaultPprofPort,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getPprofPort(tt.job))
		})
	}
}

// fakePublisher is a simple publisher that does nothing, for testing purposes.
type fakePublisher struct{}

func (f *fakePublisher) Do(_ compressor.Type, _ string, _ api.OutputType) error {
	return nil
}

func (f *fakePublisher) DoWithNativeGzipAndSplit(_, _ string, _ api.OutputType) error {
	return nil
}

// Verify that PprofProfiler result file path is generated correctly
func TestPprofProfiler_ResultFileName(t *testing.T) {
	resultFile := common.GetResultFile(common.TmpDir(), api.GoPprof, api.Raw, "pprof", 1)
	assert.Contains(t, resultFile, "raw-pprof-1.pb.gz")
}
