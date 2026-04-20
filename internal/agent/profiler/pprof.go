package profiler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/agent/config"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler/common"
	"github.com/josepdcs/kubectl-prof/internal/agent/util/publish"
	"github.com/josepdcs/kubectl-prof/pkg/util/file"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
)

const (
	// PprofHostKey is the key used in AdditionalArguments to specify the target pod's pprof host (IP).
	PprofHostKey = "pprof-host"
	// PprofPortKey is the key used in AdditionalArguments to specify the target pod's pprof port.
	PprofPortKey = "pprof-port"

	defaultPprofPort = "6060"
	defaultPprofPath = "/debug/pprof/profile"
)

// PprofProfiler is a profiler that connects to a Go application's net/http/pprof endpoint
// to retrieve profiling data over HTTP.
type PprofProfiler struct {
	PprofManager
}

// PprofManager defines the internal operations for the pprof profiler.
type PprofManager interface {
	invoke(job *job.ProfilingJob) (error, time.Duration)
}

type pprofManager struct {
	publisher  publish.Publisher
	httpClient httpClient
}

// httpClient abstracts the HTTP GET operation for testability.
type httpClient interface {
	Get(url string) (*http.Response, error)
}

// NewPprofProfiler creates a new PprofProfiler with the given publisher.
func NewPprofProfiler(publisher publish.Publisher) *PprofProfiler {
	return &PprofProfiler{
		PprofManager: &pprofManager{
			publisher:  publisher,
			httpClient: &http.Client{},
		},
	}
}

func (p *PprofProfiler) SetUp(job *job.ProfilingJob) error {
	host := getPprofHost(job)
	if host == "" {
		return errors.New("pprof host is required: set the pprof-host additional argument with the target pod IP")
	}
	log.DebugLogLn(fmt.Sprintf("Pprof profiler configured for target: %s:%s", host, getPprofPort(job)))
	return nil
}

func (p *PprofProfiler) Invoke(job *job.ProfilingJob) (error, time.Duration) {
	return p.invoke(job)
}

func (m *pprofManager) invoke(job *job.ProfilingJob) (error, time.Duration) {
	start := time.Now()

	host := getPprofHost(job)
	port := getPprofPort(job)
	seconds := strconv.Itoa(int(job.Interval.Seconds()))

	url := fmt.Sprintf("http://%s:%s%s?seconds=%s", host, port, defaultPprofPath, seconds)
	log.DebugLogLn(fmt.Sprintf("Requesting pprof data from: %s", url))

	resp, err := m.httpClient.Get(url)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve pprof data from %s", url), time.Since(start)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("pprof endpoint returned HTTP %d from %s", resp.StatusCode, url), time.Since(start)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read pprof response body"), time.Since(start)
	}

	if len(body) == 0 {
		return errors.New("pprof endpoint returned empty response"), time.Since(start)
	}

	resultFileName := common.GetResultFile(common.TmpDir(), job.Tool, job.OutputType, "pprof", job.Iteration)
	file.Write(resultFileName, string(body))

	return m.publisher.Do(job.Compressor, resultFileName, job.OutputType), time.Since(start)
}

func (p *PprofProfiler) CleanUp(*job.ProfilingJob) error {
	file.RemoveAll(common.TmpDir(), config.ProfilingPrefix)
	return nil
}

// getPprofHost retrieves the pprof host from the job's additional arguments.
func getPprofHost(job *job.ProfilingJob) string {
	if job.AdditionalArguments != nil {
		if host, ok := job.AdditionalArguments[PprofHostKey]; ok {
			return host
		}
	}
	return ""
}

// getPprofPort retrieves the pprof port from the job's additional arguments,
// defaulting to 6060 if not specified.
func getPprofPort(job *job.ProfilingJob) string {
	if job.AdditionalArguments != nil {
		if port, ok := job.AdditionalArguments[PprofPortKey]; ok && port != "" {
			return port
		}
	}
	return defaultPprofPort
}
