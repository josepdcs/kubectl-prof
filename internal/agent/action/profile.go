package action

import (
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
)

// arguments passed to the agent
const (
	JobId                      string = "job-id"
	TargetContainerRuntime            = "target-container-runtime"
	TargetContainerRuntimePath        = "target-container-runtime-path"
	TargetPodUID                      = "target-pod-uid"
	TargetContainerID                 = "target-container-id"
	Duration                          = "duration"
	Interval                          = "interval"
	Lang                              = "lang"
	EventType                         = "event-type"
	CompressorType                    = "compressor-type"
	ProfilingTool                     = "profiling-tool"
	OutputType                        = "output-type"
	Filename                          = "filename"
	PrintLogs                         = "print-logs"
	GracePeriodForEnding              = "grace-period-ending"
	OutputSplitInChunkSize            = "output-split-in-chunk-size"
	Pid                               = "pid"
	Pgrep                             = "pgrep"
	NodeHeapSnapshotSignal            = "node-heap-snapshot-signal"
	AsyncProfilerArg                  = "async-profiler-arg"
	HeartbeatInterval                 = "heartbeat-interval"
	PprofHost                         = "pprof-host"
	PprofPort                         = "pprof-port"

	defaultDuration               = 60 * time.Second
	defaultHeartbeatInterval      = 30 * time.Second
	defaultContainerRuntime       = api.Containerd
	defaultCompressor             = compressor.Gzip
	defaultEventType              = api.Ctimer
	defaultOutputSplitInChunkSize = "50M"
)

// NewProfile initializes and returns a [profiler.Profiler], [job.ProfilingJob], and any error encountered during setup.
func NewProfile(args map[string]any) (profiler.Profiler, *job.ProfilingJob, error) {
	log.SetPrintLogs(args[PrintLogs].(bool))

	profilingJob, err := getProfilingJob(args)
	if err != nil {
		return nil, nil, err
	}

	return profiler.Get(profilingJob.Tool), profilingJob, nil
}

// Run runs the profiling job using the provided [profiler.Profiler] and [job.ProfilingJob]. It returns any error encountered during execution.
func Run(p profiler.Profiler, job *job.ProfilingJob) error {
	_ = log.EventLn(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Started})

	err := p.SetUp(job)
	if err != nil {
		return err
	}

	// if Duration == Interval, one iteration occurs (discrete mode)
	iterations := int64(job.Duration.Seconds() / job.Interval.Seconds())
	var i int64
	for i = 0; i < iterations; i++ {
		job.Iteration = int(i) + 1
		err, d := p.Invoke(job)
		if err != nil {
			return err
		}
		if iterations > 1 && d.Seconds() < job.Interval.Seconds() {
			time.Sleep(time.Duration(job.Interval.Milliseconds()-d.Milliseconds()) * time.Millisecond)
		}
	}

	_ = log.EventLn(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Ended})

	return nil
}

// getProfilingJob gets a new filled config.ProfilingJob according given cli.Context
func getProfilingJob(args map[string]any) (*job.ProfilingJob, error) {
	j := &job.ProfilingJob{}

	err := validateJob(args, j)
	if err != nil {
		return nil, err
	}

	log.DebugLogLn(j.String())

	return j, nil
}
