package cmd

import (
	"fmt"
	"os"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
)

// flagValidator defines the interface for flag validators.
// It follows the Chain of Responsibility pattern.
type flagValidator interface {
	// validate executes the validation of the flags.
	validate(*profilingFlags, *config.TargetConfig, *config.JobConfig) error
	// setNext sets the next validator in the chain.
	setNext(flagValidator) flagValidator
}

// baseFlagValidator is the base structure for flag validators that
// implements common functionality for managing the next validator in the chain.
type baseFlagValidator struct {
	next flagValidator
}

// setNext sets the next validator in the chain and returns it.
func (b *baseFlagValidator) setNext(next flagValidator) flagValidator {
	b.next = next
	return next
}

// validateNext executes the validation of the next validator in the chain, if it exists.
func (b *baseFlagValidator) validateNext(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	if b.next != nil {
		return b.next.validate(flags, target, job)
	}
	return nil
}

// languageValidator validates that the programming language is supported.
type languageValidator struct {
	baseFlagValidator
}

// validate checks if the provided language is valid and assigns it to the target configuration.
func (v *languageValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	err := validateLang(flags.lang)
	if err != nil {
		return err
	}
	target.Language = api.ProgrammingLanguage(flags.lang)
	return v.validateNext(flags, target, job)
}

// runtimeValidator validates that the container runtime is supported.
type runtimeValidator struct {
	baseFlagValidator
}

// validate checks if the provided container runtime is valid and assigns it to the target configuration.
func (v *runtimeValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	var err error
	flags.runtime, err = validateRuntime(flags.runtime, target)
	if err != nil {
		return err
	}
	target.ContainerRuntime = api.ContainerRuntime(flags.runtime)
	target.ContainerRuntimePath = api.GetContainerRuntimeRootPath[target.ContainerRuntime]
	if flags.runtimePath != api.GetContainerRuntimeRootPath[api.ContainerRuntime(defaultContainerRuntime)] {
		target.ContainerRuntimePath = flags.runtimePath
	}
	return v.validateNext(flags, target, job)
}

// eventValidator validates that the profiling event is supported.
type eventValidator struct {
	baseFlagValidator
}

// validate checks if the provided profiling event is valid and assigns it to the target configuration.
func (v *eventValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	if stringUtils.IsNotBlank(flags.event) && !api.IsSupportedEvent(flags.event) {
		return errors.Errorf("unsupported event, choose one of %s", api.AvailableEvents())
	}
	if stringUtils.IsBlank(flags.event) {
		flags.event = defaultEvent
	}
	target.Event = api.ProfilingEvent(flags.event)
	return v.validateNext(flags, target, job)
}

// logLevelValidator validates that the log level is supported.
type logLevelValidator struct {
	baseFlagValidator
}

// validate checks if the provided log level is valid.
func (v *logLevelValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	if stringUtils.IsNotBlank(flags.logLevel) && !api.IsSupportedLogLevel(flags.logLevel) {
		return errors.Errorf("unsupported log level, choose one of %s", api.AvailableLogLevels())
	}
	if stringUtils.IsBlank(flags.logLevel) {
		flags.logLevel = defaultLogLevel
	}
	return v.validateNext(flags, target, job)
}

// compressorValidator validates the compressor type.
type compressorValidator struct {
	baseFlagValidator
}

// validate checks if the provided compressor type is valid and assigns it to the target configuration.
func (v *compressorValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	if stringUtils.IsBlank(flags.compressorType) {
		flags.compressorType = defaultCompressor
	}
	target.Compressor = compressor.Type(flags.compressorType)
	return v.validateNext(flags, target, job)
}

// imagePullPolicyValidator validates the image pull policy.
type imagePullPolicyValidator struct {
	baseFlagValidator
}

// validate checks if the provided image pull policy is valid and assigns it to the target configuration.
func (v *imagePullPolicyValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	if stringUtils.IsNotBlank(flags.imagePullPolicy) && !isSupportedImagePullPolicy(flags.imagePullPolicy) {
		return errors.Errorf("unsupported image pull policy, choose one of %s", imagePullPolicies)
	}
	if stringUtils.IsBlank(flags.imagePullPolicy) {
		flags.imagePullPolicy = defaultImagePullPolicy
	}
	target.ImagePullPolicy = apiv1.PullPolicy(flags.imagePullPolicy)
	return v.validateNext(flags, target, job)
}

// profilingToolAndOutputValidator validates the profiling tool and output type.
type profilingToolAndOutputValidator struct {
	baseFlagValidator
}

// validate checks if the profiling tool and output type are valid for the selected language.
func (v *profilingToolAndOutputValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	validateProfilingTool(flags.profilingTool, flags.outputType, target)
	validateOutputType(flags.outputType, target)
	return v.validateNext(flags, target, job)
}

// resourcesValidator validates requested resources, limits and tolerations for the job.
type resourcesValidator struct {
	baseFlagValidator
}

// validate checks if resource and toleration configurations are valid for the job.
func (v *resourcesValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	if _, err := job.RequestConfig.ParseResources(); err != nil {
		return errors.Wrapf(err, "unable to parse resource requests")
	}
	if _, err := job.LimitConfig.ParseResources(); err != nil {
		return errors.Wrapf(err, "unable to parse resource limits")
	}
	if err := job.ParseTolerations(); err != nil {
		return errors.Wrapf(err, "unable to parse tolerations")
	}
	return v.validateNext(flags, target, job)
}

// localPathValidator validates the local path where results will be saved.
type localPathValidator struct {
	baseFlagValidator
}

// validate checks if the local path is valid and creates it if necessary.
func (v *localPathValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	if !stringUtils.IsBlank(target.LocalPath) {
		err := os.MkdirAll(target.LocalPath, 0755)
		if err != nil {
			return errors.Wrap(err, "could not create local path")
		}
	}
	return v.validateNext(flags, target, job)
}

// pidValidator validates the process ID (PID).
type pidValidator struct {
	baseFlagValidator
}

// validate checks if the provided PID is numeric.
func (v *pidValidator) validate(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	return validatePid(target.PID)
}

// validateFlags orchestrates the validation of all flags using a chain of validators.
func validateFlags(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	validator := &languageValidator{}
	validator.setNext(&runtimeValidator{}).
		setNext(&eventValidator{}).
		setNext(&logLevelValidator{}).
		setNext(&compressorValidator{}).
		setNext(&imagePullPolicyValidator{}).
		setNext(&profilingToolAndOutputValidator{}).
		setNext(&resourcesValidator{}).
		setNext(&localPathValidator{}).
		setNext(&pidValidator{})

	return validator.validate(flags, target, job)
}

// validatePid checks if the provided PID is numeric.
func validatePid(pid string) error {
	if !stringUtils.IsNumeric(pid) {
		return errors.New("pid must be numeric")
	}
	return nil
}

// validateRuntime checks if the container runtime is supported.
// If none is provided, the default one is used.
func validateRuntime(runtime string, target *config.TargetConfig) (string, error) {
	if stringUtils.IsNotBlank(runtime) && !api.IsSupportedContainerRuntime(runtime) {
		return "", errors.Errorf("unsupported container runtime, choose one of %s", api.AvailableContainerRuntimes())
	}
	if stringUtils.IsBlank(runtime) {
		runtime = defaultContainerRuntime
		target.ContainerRuntimePath = api.GetContainerRuntimeRootPath[api.ContainerRuntime(defaultContainerRuntime)]
	}
	return runtime, nil
}

// validateLang checks if the programming language is supported.
func validateLang(lang string) error {
	if lang == "" {
		return errors.Errorf("use -l flag to select one of the supported languages %s", api.AvailableLanguages())
	}

	if !api.IsSupportedLanguage(lang) {
		return errors.Errorf("unsupported language, choose one of %s", api.AvailableLanguages())
	}
	return nil
}

// validateProfilingTool checks if the provided profiling tool is valid for the selected language.
// If none is provided, the default one is used.
func validateProfilingTool(profilingTool string, outputType string, target *config.TargetConfig) {
	if stringUtils.IsBlank(profilingTool) {
		target.ProfilingTool = api.GetProfilingTool(target.Language, api.OutputType(outputType))
		fmt.Printf("Default profiling tool %s will be used ... 🧐\n", target.ProfilingTool)
		return
	}

	defaultTool := api.GetProfilingToolsByProgrammingLanguage[target.Language][0]
	if !api.IsSupportedProfilingTool(profilingTool) {
		fmt.Printf("Unsupported profiling tool %s, default %s will be used ... 🧐\n", profilingTool, defaultTool)
		target.ProfilingTool = defaultTool
		return
	}

	if !api.IsValidProfilingTool(api.ProfilingTool(profilingTool), target.Language) {
		fmt.Printf("Unsupported profiling tool %s for language %s, default %s will be used ... 🧐\n",
			profilingTool, target.Language, defaultTool)
		target.ProfilingTool = defaultTool
		return
	}

	target.ProfilingTool = api.ProfilingTool(profilingTool)
}

// validateOutputType checks if the provided output type is valid for the selected profiling tool.
// If none is provided, the default one is used.
func validateOutputType(outputType string, target *config.TargetConfig) {
	defaultOutputType := api.GetOutputTypesByProfilingTool[target.ProfilingTool][0]
	if outputType == "" {
		fmt.Printf("Default output type %s will be used ... 🧐\n", defaultOutputType)
		target.OutputType = defaultOutputType
		return
	}

	if !api.IsSupportedOutputType(outputType) {
		fmt.Printf("Unsupported output type %s, default %s will be used ... ✔\n", outputType, defaultOutputType)
		target.OutputType = defaultOutputType
		return
	}

	if !api.IsValidOutputType(api.OutputType(outputType), target.ProfilingTool) {
		fmt.Printf("Unsupported output type %s for profiling tool %s, default %s will be used ... ✔\n",
			outputType, target.ProfilingTool, defaultOutputType)
		target.OutputType = defaultOutputType
		return
	}

	target.OutputType = api.OutputType(outputType)
}

// isSupportedImagePullPolicy checks if the image pull policy is supported.
func isSupportedImagePullPolicy(imagePullPolicy string) bool {
	for _, current := range imagePullPolicies {
		if apiv1.PullPolicy(imagePullPolicy) == current {
			return true
		}
	}
	return false
}
