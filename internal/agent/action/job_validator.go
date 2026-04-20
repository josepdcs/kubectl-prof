package action

import (
	"fmt"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/agent/job"
	"github.com/josepdcs/kubectl-prof/internal/agent/util"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/josepdcs/kubectl-prof/pkg/util/log"
	"github.com/pkg/errors"
)

// jobValidator defines the interface for job configuration validators and fillers.
// It follows the Chain of Responsibility pattern.
type jobValidator interface {
	// validate executes the validation and filling of the profiling job.
	validate(map[string]any, *job.ProfilingJob) error
	// setNext sets the next validator in the chain.
	setNext(jobValidator) jobValidator
}

// baseJobValidator is the base structure for job validators that
// implements common functionality for managing the next validator in the chain.
type baseJobValidator struct {
	next jobValidator
}

// setNext sets the next validator in the chain and returns it.
func (b *baseJobValidator) setNext(next jobValidator) jobValidator {
	b.next = next
	return next
}

// validateNext executes the validation of the next validator in the chain, if it exists.
func (b *baseJobValidator) validateNext(args map[string]any, j *job.ProfilingJob) error {
	if b.next != nil {
		return b.next.validate(args, j)
	}
	return nil
}

// durationIntervalValidator validates and sets the duration and interval for the profiling job.
type durationIntervalValidator struct {
	baseJobValidator
}

// validate checks and sets the duration and interval.
func (v *durationIntervalValidator) validate(args map[string]any, j *job.ProfilingJob) error {
	j.Duration = defaultDuration
	if stringUtils.IsNotBlank(args[Duration].(string)) {
		duration, err := time.ParseDuration(args[Duration].(string))
		if err != nil {
			return err
		}
		j.Duration = duration
	}

	if stringUtils.IsNotBlank(args[Interval].(string)) {
		interval, err := time.ParseDuration(args[Interval].(string))
		if err != nil {
			return err
		}
		j.Interval = interval
	} else {
		j.Interval = j.Duration
	}

	if j.Interval > j.Duration {
		return errors.Errorf("interval cannot be greater than duration (duration %d, interval: %d)", j.Duration, j.Interval)
	}
	return v.validateNext(args, j)
}

// containerRuntimeValidator validates and sets the container runtime and its path.
type containerRuntimeValidator struct {
	baseJobValidator
}

// validate checks and sets the container runtime and its path.
func (v *containerRuntimeValidator) validate(args map[string]any, j *job.ProfilingJob) error {
	containerRuntime := args[TargetContainerRuntime].(string)
	if stringUtils.IsBlank(containerRuntime) {
		j.ContainerRuntime = defaultContainerRuntime
	} else if !api.IsSupportedContainerRuntime(containerRuntime) {
		return errors.Errorf("unsupported container runtime, choose one of %s", api.AvailableContainerRuntimes())
	} else {
		j.ContainerRuntime = api.ContainerRuntime(containerRuntime)
	}

	containerRuntimePath := args[TargetContainerRuntimePath].(string)
	if stringUtils.IsBlank(containerRuntimePath) {
		j.ContainerRuntimePath = api.GetContainerRuntimeRootPath[j.ContainerRuntime]
	} else {
		j.ContainerRuntimePath = containerRuntimePath
	}
	return v.validateNext(args, j)
}

// identificationValidator sets the various IDs for the profiling job.
type identificationValidator struct {
	baseJobValidator
}

// validate sets UID, PodUID, ContainerID, and FileName.
func (v *identificationValidator) validate(args map[string]any, j *job.ProfilingJob) error {
	j.UID = args[JobId].(string)
	j.PodUID = args[TargetPodUID].(string)
	j.ContainerID = util.NormalizeContainerID(args[TargetContainerID].(string))
	j.FileName = args[Filename].(string)
	return v.validateNext(args, j)
}

// languageValidator validates and sets the programming language.
type languageValidator struct {
	baseJobValidator
}

// validate checks and sets the programming language.
func (v *languageValidator) validate(args map[string]any, j *job.ProfilingJob) error {
	lang := args[Lang].(string)
	if !api.IsSupportedLanguage(lang) {
		return errors.Errorf("unsupported language, choose one of %s", api.AvailableLanguages())
	}
	j.Language = api.ProgrammingLanguage(lang)
	return v.validateNext(args, j)
}

// eventValidator validates and sets the profiling event.
type eventValidator struct {
	baseJobValidator
}

// validate checks and sets the profiling event.
func (v *eventValidator) validate(args map[string]any, j *job.ProfilingJob) error {
	event := args[EventType].(string)
	if stringUtils.IsBlank(event) {
		event = string(defaultEventType)
	}
	if !api.IsSupportedEvent(event) {
		return errors.Errorf("unsupported event, choose one of %s", api.AvailableEvents())
	}
	j.Event = api.ProfilingEvent(event)
	return v.validateNext(args, j)
}

// compressorValidator validates and sets the compressor type.
type compressorValidator struct {
	baseJobValidator
}

// validate checks and sets the compressor type.
func (v *compressorValidator) validate(args map[string]any, j *job.ProfilingJob) error {
	co := args[CompressorType].(string)
	if stringUtils.IsBlank(co) {
		co = string(defaultCompressor)
	}
	if !compressor.IsSupportedCompressor(co) {
		return errors.Errorf("unsupported compressor, choose one of %s", compressor.AvailableCompressors())
	}
	j.Compressor = compressor.Type(co)
	return v.validateNext(args, j)
}

// profilingToolAndOutputValidator validates and sets the profiling tool and output type.
type profilingToolAndOutputValidator struct {
	baseJobValidator
}

// validate checks and sets the profiling tool and output type.
func (v *profilingToolAndOutputValidator) validate(args map[string]any, j *job.ProfilingJob) error {
	validateProfilingTool(args[ProfilingTool].(string), args[OutputType].(string), j)
	validateOutputType(args[OutputType].(string), j)
	return v.validateNext(args, j)
}

// additionalParametersValidator sets optional parameters for the profiling job.
type additionalParametersValidator struct {
	baseJobValidator
}

// validate sets various additional parameters if they are provided.
func (v *additionalParametersValidator) validate(args map[string]any, j *job.ProfilingJob) error {
	setOutputSplitChunkSize(args, j)
	setPid(args, j)
	setPgrep(args, j)
	setNodeHeapSnapshotSignal(args, j)
	setAsyncProfilerArgs(args, j)
	setHeartbeatInterval(args, j)
	setPprofArgs(args, j)
	return v.validateNext(args, j)
}

// validateJob orchestrates the validation and filling of the profiling job using a chain of validators.
func validateJob(args map[string]any, j *job.ProfilingJob) error {
	validator := &durationIntervalValidator{}
	validator.setNext(&containerRuntimeValidator{}).
		setNext(&identificationValidator{}).
		setNext(&languageValidator{}).
		setNext(&eventValidator{}).
		setNext(&compressorValidator{}).
		setNext(&profilingToolAndOutputValidator{}).
		setNext(&additionalParametersValidator{})

	return validator.validate(args, j)
}

// setOutputSplitChunkSize sets the output split chunk size for relevant output types.
func setOutputSplitChunkSize(args map[string]any, j *job.ProfilingJob) {
	if j.OutputType == api.HeapDump || j.OutputType == api.HeapSnapshot ||
		j.OutputType == api.Gcdump || j.OutputType == api.Dump {
		j.OutputSplitInChunkSize = defaultOutputSplitInChunkSize
		if args[OutputSplitInChunkSize] != nil && stringUtils.IsNotBlank(args[OutputSplitInChunkSize].(string)) {
			j.OutputSplitInChunkSize = args[OutputSplitInChunkSize].(string)
		}
	}
}

// setPid sets the process ID (PID) if provided.
func setPid(args map[string]any, j *job.ProfilingJob) {
	if args[Pid] != nil && stringUtils.IsNotBlank(args[Pid].(string)) {
		j.PID = args[Pid].(string)
	}
}

// setPgrep sets the pgrep filter if provided.
func setPgrep(args map[string]any, j *job.ProfilingJob) {
	if args[Pgrep] != nil && stringUtils.IsNotBlank(args[Pgrep].(string)) {
		j.Pgrep = args[Pgrep].(string)
	}
}

// setNodeHeapSnapshotSignal sets the Node.js heap snapshot signal if provided.
func setNodeHeapSnapshotSignal(args map[string]any, j *job.ProfilingJob) {
	if args[NodeHeapSnapshotSignal] != nil {
		j.NodeHeapSnapshotSignal = args[NodeHeapSnapshotSignal].(int)
	}
}

// setAsyncProfilerArgs sets additional arguments for async-profiler if provided.
func setAsyncProfilerArgs(args map[string]any, j *job.ProfilingJob) {
	if args[AsyncProfilerArg] != nil {
		asyncProfilerArgs := args[AsyncProfilerArg].([]string)
		if len(asyncProfilerArgs) > 0 {
			if j.AdditionalArguments == nil {
				j.AdditionalArguments = make(map[string]string)
			}
			// Store each async-profiler argument with an indexed key (async-profiler-arg-0, async-profiler-arg-1, etc.)
			// These will be retrieved in order when building the async-profiler command
			for i, arg := range asyncProfilerArgs {
				key := fmt.Sprintf("async-profiler-arg-%d", i)
				j.AdditionalArguments[key] = arg
			}
		}
	}
}

// setHeartbeatInterval sets the heartbeat interval for profilers that need periodic progress events.
func setHeartbeatInterval(args map[string]any, j *job.ProfilingJob) {
	j.HeartbeatInterval = defaultHeartbeatInterval
	if args[HeartbeatInterval] != nil && stringUtils.IsNotBlank(args[HeartbeatInterval].(string)) {
		interval, err := time.ParseDuration(args[HeartbeatInterval].(string))
		if err == nil && interval > 0 {
			j.HeartbeatInterval = interval
		}
	}
}

// validateProfilingTool validates the given profiling tool and sets the default tool if needed.
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

// validateOutputType validates the given output type and sets the default output type if needed.
func validateOutputType(outputType string, job *job.ProfilingJob) {
	defaultOutputType := api.GetOutputTypesByProfilingTool[job.Tool][0]
	if stringUtils.IsBlank(outputType) {
		job.OutputType = defaultOutputType
		log.InfoLogLn(fmt.Sprintf("Default output type %s will be used", job.OutputType))
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

// setPprofArgs sets the pprof host and port from the CLI args into the job's additional arguments.
func setPprofArgs(args map[string]any, j *job.ProfilingJob) {
	if args[PprofHost] != nil && stringUtils.IsNotBlank(args[PprofHost].(string)) {
		if j.AdditionalArguments == nil {
			j.AdditionalArguments = make(map[string]string)
		}
		j.AdditionalArguments["pprof-host"] = args[PprofHost].(string)
	}
	if args[PprofPort] != nil && stringUtils.IsNotBlank(args[PprofPort].(string)) {
		if j.AdditionalArguments == nil {
			j.AdditionalArguments = make(map[string]string)
		}
		j.AdditionalArguments["pprof-port"] = args[PprofPort].(string)
	}
}
