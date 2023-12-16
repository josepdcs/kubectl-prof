package cmd

import (
	"fmt"
	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/internal/cli"
	"github.com/josepdcs/kubectl-prof/internal/cli/adapter"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/profiler"
	"github.com/josepdcs/kubectl-prof/internal/cli/version"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"os"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	defaultGracePeriodEnding      = 5 * time.Minute
	defaultContainerRuntime       = string(api.Containerd)
	defaultEvent                  = string(api.Itimer)
	defaultLogLevel               = string(api.InfoLevel)
	defaultCompressor             = string(compressor.Gzip)
	defaultOutputType             = string(api.FlameGraph)
	defaultImagePullPolicy        = string(apiv1.PullIfNotPresent)
	defaultHeapDumpSplitSize      = "50M"
	defaultPoolSizeRetrieveChunks = 5
	defaultRetrieveFileRetries    = 3
	longDescription               = `Profiling on existing applications with low-overhead.

These commands help you identify application performance issues.
`
	profilingExamples = `
	# Profile a pod for 5 minutes with JFR format for java language
	%[1]s prof my-pod -t 5m -l java -o jfr

	# Profile an alpine based container for java language
	%[1]s prof mypod -l java --alpine 

	# Profile a pod for 5 minutes in intervals of 60 seconds for java language by giving the cpu limits, the container runtime, the agent image and the image pull policy
	%[1]s my-pod -l java -o flamegraph -t 5m --interval 60s --cpu-limits=1 -r containerd --image=localhost/my-agent-image-jvm:latest --image-pull-policy=IfNotPresent

	# Profile in contprof namespace a pod running in contprof-stupid-apps namespace by using the profiler service account for go language 
	%[1]s prof mypod -n contprof --service-account=profiler --target-namespace=contprof-stupid-apps -l go

	# Set custom resource requests and limits for the agent pod (default: neither requests nor limits are set) for python language
	%[1]s prof mypod --cpu-requests 100m --cpu-limits 200m --mem-requests 100Mi --mem-limits 200Mi -l python
`
)

var imagePullPolicies = []apiv1.PullPolicy{apiv1.PullNever, apiv1.PullAlways, apiv1.PullIfNotPresent}

// Profiler defines the profile method.
type Profiler interface {
	Profile(cfg *config.ProfilerConfig) error
}

type ProfileOptions struct {
	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams
}

func NewProfileOptions(streams genericclioptions.IOStreams) *ProfileOptions {
	return &ProfileOptions{
		configFlags: genericclioptions.NewConfigFlags(false),
		IOStreams:   streams,
	}
}

func NewProfileCommand(streams genericclioptions.IOStreams) *cobra.Command {
	var (
		target                config.TargetConfig
		job                   config.JobConfig
		ephemeralContainer    config.EphemeralContainerConfig
		showVersion           bool
		runtime               string
		lang                  string
		event                 string
		logLevel              string
		compressorType        string
		profilingTool         string
		outputType            string
		imagePullPolicy       string
		privileged            bool
		useEphemeralContainer = false // disabled
	)

	options := NewProfileOptions(streams)
	cmd := &cobra.Command{
		Use:                   "prof [pod-name]",
		DisableFlagsInUseLine: true,
		Short:                 "Profile running applications. Several output types are supported: flamegraphs, jfrs, threadumps, heapdumps, etc.",
		Long:                  longDescription,
		Example:               fmt.Sprintf(profilingExamples, "kubectl"),
		PersistentPreRun: func(c *cobra.Command, args []string) {
			c.SetOut(streams.Out)
			c.SetErr(streams.ErrOut)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				_, _ = fmt.Fprintln(streams.Out, version.String())
				return
			}

			if len(args) == 0 {
				_ = cmd.Help()
				return
			}

			if err := validateFlags(runtime, lang, event, logLevel, compressorType, profilingTool, outputType, imagePullPolicy, &target, &job); err != nil {
				_, _ = fmt.Fprintln(streams.Out, err)
				os.Exit(1)
			}

			// set log level
			level, _ := log.ParseLevel(logLevel)
			log.SetLevel(level)

			target.PodName = args[0]
			if len(args) > 1 {
				target.ContainerName = args[1]
			}

			// Prepare profiler
			cfg, err := getProfilerConfig(target, job, ephemeralContainer, useEphemeralContainer, logLevel, privileged)
			if err != nil {
				log.Fatalf("Failed configure profiler: %v\n", err)
			}

			connectionInfo, err := kubernetes.Connect(options.configFlags)
			if err != nil {
				log.Fatalf("Failed connecting to kubernetes cluster: %v\n", err)
			}

			if cfg.Target.Namespace == "" {
				cfg.Target.Namespace = connectionInfo.Namespace
			}

			if useEphemeralContainer {
				err = profiler.NewEphemeralProfiler(
					adapter.NewPodAdapter(connectionInfo),
					adapter.NewProfilingEphemeralContainerAdapter(connectionInfo),
					adapter.NewProfilingContainerAdapter(connectionInfo),
				).Profile(cfg)
			} else {
				cfg.Job.Namespace = connectionInfo.Namespace
				err = profiler.NewJobProfiler(
					adapter.NewPodAdapter(connectionInfo),
					adapter.NewProfilingJobAdapter(connectionInfo),
					adapter.NewProfilingContainerAdapter(connectionInfo),
				).Profile(cfg)
			}

			if err != nil {
				printer := cli.NewPrinter(cfg.Target.DryRun)
				printer.PrintError()
				log.Fatalf(err.Error())
			}
		},
	}

	cmd.Flags().BoolVar(&showVersion, "version", false, "Print version info")

	cmd.Flags().StringVarP(&runtime, "runtime", "r", defaultContainerRuntime,
		fmt.Sprintf("The container runtime used for kubernetes, choose one of %v", api.AvailableContainerRuntimes()))
	cmd.Flags().StringVar(&target.ContainerRuntimePath, "runtime-path", api.GetContainerRuntimeRootPath[api.ContainerRuntime(defaultContainerRuntime)],
		"Use a different container runtime install path")

	cmd.Flags().DurationVarP(&target.Duration, "time", "t", 0, "Max scan Duration")
	// if interval is not given, duration is set as default
	cmd.Flags().DurationVar(&target.Interval, "interval", target.Duration, "Max scan Interval")
	cmd.Flags().StringVar(&target.LocalPath, "local-path", "", "Optional local path location to store the result files. Default is current directory")
	cmd.Flags().BoolVar(&target.Alpine, "alpine", false, "TargetConfig image is based on Alpine")
	cmd.Flags().BoolVar(&target.DryRun, "dry-run", false, "Simulate profiling")
	cmd.Flags().StringVar(&target.Image, "image", "", "Manually choose agent docker image")
	cmd.Flags().StringVar(&target.Namespace, "target-namespace", "", "namespace of target pod if different from job namespace")

	cmd.Flags().StringVarP(&lang, "lang", "l", "",
		fmt.Sprintf("Programming language of the target application, choose one of %v", api.AvailableLanguages()))
	cmd.Flags().StringVarP(&event, "event", "e", defaultEvent,
		fmt.Sprintf("Profiling event, choose one of %v", api.AvailableEvents()))

	cmd.Flags().StringVar(&job.RequestConfig.CPU, "cpu-requests", "", "CPU requests of the started profiling container")
	cmd.Flags().StringVar(&job.RequestConfig.Memory, "mem-requests", "", "Memory requests of the started profiling container")
	cmd.Flags().StringVar(&job.LimitConfig.CPU, "cpu-limits", "", "CPU limits of the started profiling container")
	cmd.Flags().StringVar(&job.LimitConfig.Memory, "mem-limits", "", "Memory limits of the started profiling container")
	cmd.Flags().StringVar(&target.ImagePullSecret, "image-pull-secret", "", "imagePullSecret for agent docker image")
	cmd.Flags().StringVar(&target.ServiceAccountName, "service-account", "", "serviceAccountName to be used for profiling container")

	cmd.Flags().BoolVar(&privileged, "privileged", true, "Run agent container in privileged mode")
	cmd.Flags().StringVar(&logLevel, "log-level", defaultLogLevel,
		fmt.Sprintf("Log level, choose one of %v", api.AvailableLogLevels()))
	/*cmd.Flags().StringVarP(&compressorType, "compressor", "c", defaultCompressor,
	fmt.Sprintf("Compressor for compressing generated profiling result, choose one of %v", compressor.AvailableCompressors()))*/
	cmd.Flags().StringVar(&profilingTool, "tool", "", fmt.Sprintf("Profiling tool, choose one accorfing language %v", api.AvailableProfilingToolsString()))
	cmd.Flags().StringVarP(&outputType, "output", "o", defaultOutputType,
		fmt.Sprintf("Output type, choose one accorting tool %v", api.AvailableOutputTypesString()))
	cmd.Flags().BoolVar(&target.PrintLogs, "print-logs", true, "Force agent to print the log messages type to standard output")
	cmd.Flags().DurationVar(&target.GracePeriodEnding, "grace-period-ending", defaultGracePeriodEnding, "The grace period to spend before to end the agent")
	cmd.Flags().StringVar(&imagePullPolicy, "image-pull-policy", defaultImagePullPolicy, fmt.Sprintf("Image pull policy, choose one of %v", imagePullPolicies))
	cmd.Flags().StringVar(&target.ContainerName, "target-container-name", "", "The target container name to be profiled")
	cmd.Flags().StringVar(&target.HeapDumpSplitInChunkSize, "heap-dump-split-size", defaultHeapDumpSplitSize, "The heap dump will be split into chunks of given size following the split command valid format.")
	cmd.Flags().IntVar(&target.PoolSizeRetrieveChunks, "pool-size-retrieve-chunks", defaultPoolSizeRetrieveChunks, "The pool size of go routines to retrieve chunks of the obtained heap dump from the agent.")
	cmd.Flags().IntVar(&target.RetrieveFileRetries, "retrieve-file-retries", defaultRetrieveFileRetries, "The number of retries to retrieve a file from the remote container.")
	cmd.Flags().StringVar(&target.PID, "pid", "", "The PID of target process if it is known")
	//cmd.Flags().StringVarP(&target.Pgrep, "pgrep", "p", "", "Name of the target process")
	//cmd.Flags().BoolVar(&useEphemeralContainer, "use-ephemeral-container", false, "Launching profiling agent into ephemeral container instead into Job (experimental)")

	options.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func getProfilerConfig(target config.TargetConfig, job config.JobConfig, ephemeralContainer config.EphemeralContainerConfig,
	useEphemeralContainer bool, logLevel string, privileged bool) (*config.ProfilerConfig, error) {
	if useEphemeralContainer {
		ephemeralContainer = config.EphemeralContainerConfig{Privileged: privileged}
		return config.NewProfilerConfig(&target, config.WithEphemeralContainer(&ephemeralContainer), config.WithLogLevel(api.LogLevel(logLevel)))
	}

	job.Privileged = privileged
	return config.NewProfilerConfig(&target, config.WithJob(&job), config.WithLogLevel(api.LogLevel(logLevel)))
}

func validateFlags(runtime string, lang string, event string, logLevel string, compressorType string, profilingTool string,
	outputType string, imagePullPolicy string, target *config.TargetConfig, job *config.JobConfig) error {
	var err error

	err = validateLang(lang)
	if err != nil {
		return err
	}

	runtime, err = validateRuntime(runtime, target)
	if err != nil {
		return err
	}

	if stringUtils.IsNotBlank(event) && !api.IsSupportedEvent(event) {
		return errors.Errorf("unsupported event, choose one of %s", api.AvailableEvents())
	}
	if stringUtils.IsBlank(event) {
		event = defaultEvent
	}

	if stringUtils.IsNotBlank(logLevel) && !api.IsSupportedLogLevel(logLevel) {
		return errors.Errorf("unsupported log level, choose one of %s", api.AvailableLogLevels())
	}
	if stringUtils.IsBlank(logLevel) {
		logLevel = defaultLogLevel
	}

	/*if stringUtils.IsNotBlank(compressorType) && !compressor.IsSupportedCompressor(compressorType) {
		return errors.Errorf("unsupported compressor, choose one of %s", compressor.AvailableCompressors())
	}*/
	if stringUtils.IsBlank(compressorType) {
		compressorType = defaultCompressor
	}

	if stringUtils.IsNotBlank(imagePullPolicy) && !isSupportedImagePullPolicy(imagePullPolicy) {
		return errors.Errorf("unsupported image pull policy, choose one of %s", imagePullPolicies)
	}
	if stringUtils.IsBlank(imagePullPolicy) {
		imagePullPolicy = defaultImagePullPolicy
	}

	target.ImagePullPolicy = apiv1.PullPolicy(imagePullPolicy)
	target.Language = api.ProgrammingLanguage(lang)
	target.ContainerRuntime = api.ContainerRuntime(runtime)
	target.ContainerRuntimePath = api.GetContainerRuntimeRootPath[target.ContainerRuntime]
	target.Event = api.ProfilingEvent(event)
	target.Compressor = compressor.Type(compressorType)

	validateProfilingTool(profilingTool, outputType, target)
	validateOutputType(outputType, target)

	if _, err := job.RequestConfig.ParseResources(); err != nil {
		return errors.Wrapf(err, "unable to parse resource requests")
	}

	if _, err := job.LimitConfig.ParseResources(); err != nil {
		return errors.Wrapf(err, "unable to parse resource limits")
	}

	return validatePid(target.PID)
}

func validatePid(pid string) error {
	if !stringUtils.IsNumeric(pid) {
		return errors.New("pid must be numeric")
	}
	return nil
}

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

func validateLang(lang string) error {
	if lang == "" {
		return errors.Errorf("use -l flag to select one of the supported languages %s", api.AvailableLanguages())
	}

	if !api.IsSupportedLanguage(lang) {
		return errors.Errorf("unsupported language, choose one of %s", api.AvailableLanguages())
	}
	return nil
}

func validateProfilingTool(profilingTool string, outputType string, target *config.TargetConfig) {
	if stringUtils.IsBlank(profilingTool) {
		target.ProfilingTool = api.GetProfilingTool(target.Language, api.OutputType(outputType))
		fmt.Printf("Default profiling tool %s will be used ... ✔\n", target.ProfilingTool)
		return
	}

	defaultTool := api.GetProfilingToolsByProgrammingLanguage[target.Language][0]
	if !api.IsSupportedProfilingTool(profilingTool) {
		fmt.Printf("Unsupported profiling tool %s, default %s will be used ... ✔\n", profilingTool, defaultTool)
		target.ProfilingTool = defaultTool
		return
	}

	if !api.IsValidProfilingTool(api.ProfilingTool(profilingTool), target.Language) {
		fmt.Printf("Unsupported profiling tool %s for language %s, default %s will be used ... ✔\n",
			profilingTool, target.Language, defaultTool)
		target.ProfilingTool = defaultTool
		return
	}

	target.ProfilingTool = api.ProfilingTool(profilingTool)
}

func validateOutputType(outputType string, target *config.TargetConfig) {
	defaultOutputType := api.GetOutputTypesByProfilingTool[target.ProfilingTool][0]
	if outputType == "" {
		fmt.Printf("Default output type %s will be used ... ✔\n", defaultOutputType)
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

func isSupportedImagePullPolicy(imagePullPolicy string) bool {
	for _, current := range imagePullPolicies {
		if apiv1.PullPolicy(imagePullPolicy) == current {
			return true
		}
	}
	return false
}
