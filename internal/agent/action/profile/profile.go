package profile

import (
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/profiler"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"time"
)

// arguments passed to the agent
const (
	JobId                    string = "job-id"
	TargetContainerRuntime          = "target-container-runtime"
	TargetPodUID                    = "target-pod-uid"
	TargetContainerID               = "target-container-id"
	Duration                        = "duration"
	Interval                        = "interval"
	Lang                            = "lang"
	EventType                       = "event-type"
	CompressorType                  = "compressor-type"
	ProfilingTool                   = "profiling-tool"
	OutputType                      = "output-type"
	Filename                        = "filename"
	PrintLogs                       = "print-logs"
	GracePeriodForEnding            = "grace-period-ending"
	HeapDumpSplitInChunkSize        = "heap-dump-split-in-chunk-size"

	defaultDuration                 = 60 * time.Second
	defaultContainerRuntime         = api.Containerd
	defaultCompressor               = compressor.Gzip
	defaultEventType                = api.Itimer
	defaultHeapDumpSplitInChunkSize = "50M"
)

func NewAction(args map[string]interface{}) (profiler.Profiler, *job.ProfilingJob, error) {
	log.SetPrintLogs(args[PrintLogs].(bool))

	profilingJob, err := getProfilingJob(args)
	if err != nil {
		return nil, nil, err
	}

	return profiler.Get(profilingJob.Tool), profilingJob, nil
}

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
func getProfilingJob(args map[string]interface{}) (*job.ProfilingJob, error) {
	j := &job.ProfilingJob{}

	// duration is set as mandatory
	j.Duration = defaultDuration
	if stringUtils.IsNotBlank(args[Duration].(string)) {
		duration, err := time.ParseDuration(args[Duration].(string))
		if err != nil {
			return nil, err
		}
		j.Duration = duration
	}

	// if interval is not given, duration is taken as interval value
	if stringUtils.IsNotBlank(args[Interval].(string)) {
		interval, err := time.ParseDuration(args[Interval].(string))
		if err != nil {
			return nil, err
		}
		j.Interval = interval
	} else {
		j.Interval = j.Duration
	}

	if j.Interval > j.Duration {
		return nil, fmt.Errorf("interval cannot be greater than duration (duration %d, interval: %d)", j.Duration, j.Interval)
	}

	containerRuntime := args[TargetContainerRuntime].(string)
	if stringUtils.IsBlank(containerRuntime) {
		j.ContainerRuntime = defaultContainerRuntime
	}
	if !api.IsSupportedContainerRuntime(containerRuntime) {
		return nil, fmt.Errorf("unsupported container runtime, choose one of %s", api.AvailableContainerRuntimes())
	}
	j.ContainerRuntime = api.ContainerRuntime(containerRuntime)
	j.UID = args[JobId].(string)
	j.PodUID = args[TargetPodUID].(string)
	j.ContainerID = util.NormalizeContainerID(args[TargetContainerID].(string))
	j.FileName = args[Filename].(string)

	// TODO improve validations (maybe applying the chain of responsibility pattern)
	lang := args[Lang].(string)
	if !api.IsSupportedLanguage(lang) {
		return nil, fmt.Errorf("unsupported language, choose one of %s", api.AvailableLanguages())
	}
	j.Language = api.ProgrammingLanguage(lang)

	event := args[EventType].(string)
	if stringUtils.IsBlank(event) {
		event = string(defaultEventType)
	}
	if !api.IsSupportedEvent(event) {
		return nil, fmt.Errorf("unsupported event, choose one of %s", api.AvailableEvents())
	}
	j.Event = api.ProfilingEvent(event)

	co := args[CompressorType].(string)
	if stringUtils.IsBlank(co) {
		co = defaultCompressor
	}
	if !compressor.IsSupportedCompressor(co) {
		return nil, fmt.Errorf("unsupported compressor, choose one of %s", compressor.AvailableCompressors())
	}
	j.Compressor = compressor.Type(co)

	validateProfilingTool(args[ProfilingTool].(string), args[OutputType].(string), j)
	validateOutputType(args[OutputType].(string), j)

	// set heap dump split in chunk size
	if j.OutputType == api.HeapDump {
		j.HeapDumpSplitInChunkSize = defaultHeapDumpSplitInChunkSize
		if args[HeapDumpSplitInChunkSize] != nil && stringUtils.IsNotBlank(args[HeapDumpSplitInChunkSize].(string)) {
			j.HeapDumpSplitInChunkSize = args[HeapDumpSplitInChunkSize].(string)
		}
	}

	log.DebugLogLn(j.String())

	return j, nil
}

// validateProfilingTool validates the given profiling tool and sets the default tool if needed
func validateProfilingTool(profilingTool string, outputType string, job *job.ProfilingJob) {
	if stringUtils.IsBlank(profilingTool) {
		job.Tool = api.GetProfilingTool(job.Language, api.OutputType(outputType))
		log.InfoLogLn(fmt.Sprintf("Default profiling tool %s will be used", job.Tool))
		return
	}

	defaultTool := api.GetProfilingToolsByProgrammingLanguage[job.Language][0]
	if !api.IsSupportedProfilingTool(profilingTool) {
		log.WarningLogLn(fmt.Sprintf("Unsupported profiling tool %s, default %s will be used",
			profilingTool, defaultTool))
		job.Tool = defaultTool
		return
	}

	if !api.IsValidProfilingTool(api.ProfilingTool(profilingTool), job.Language) {
		log.WarningLogLn(fmt.Sprintf("Unsupported profiling tool %s for language %s, default %s will be used",
			profilingTool, job.Language, defaultTool))
		job.Tool = defaultTool
		return
	}

	job.Tool = api.ProfilingTool(profilingTool)
}

// validateOutputType validates the given profiling tool and sets the default tool if needed
func validateOutputType(outputType string, job *job.ProfilingJob) {
	defaultOutputType := api.GetOutputTypesByProfilingTool[job.Tool][0]
	if stringUtils.IsBlank(outputType) {
		log.WarningLogLn(fmt.Sprintf("Default output type %s will be used", defaultOutputType))
		job.OutputType = defaultOutputType
		return
	}

	if !api.IsSupportedOutputType(outputType) {
		log.WarningLogLn(fmt.Sprintf("Unsupported output type %s, default %s will be used",
			outputType, defaultOutputType))
		job.OutputType = defaultOutputType
		return
	}

	if !api.IsValidOutputType(api.OutputType(outputType), job.Tool) {
		log.WarningLogLn(fmt.Sprintf("Unsupported output type %s for profiling tool %s, default %s will be used",
			outputType, job.Tool, defaultOutputType))
		job.OutputType = defaultOutputType
		return
	}

	job.OutputType = api.OutputType(outputType)
}
