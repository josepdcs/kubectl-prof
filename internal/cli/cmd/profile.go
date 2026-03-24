package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/josepdcs/kubectl-prof/internal/cli"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/profiler"
	apiprof "github.com/josepdcs/kubectl-prof/internal/cli/profiler/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/version"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	defaultGracePeriodEnding           = 5 * time.Minute
	defaultContainerRuntime            = string(api.Containerd)
	defaultEvent                       = string(api.Ctimer)
	defaultLogLevel                    = string(api.InfoLevel)
	defaultCompressor                  = string(compressor.Gzip)
	defaultOutputType                  = string(api.FlameGraph)
	defaultImagePullPolicy             = string(apiv1.PullIfNotPresent)
	defaultPoolSizeLaunchProfilingJobs = 0
	defaultOutputSplitSize             = "50M"
	defaultPoolSizeRetrieveChunks      = 5
	defaultRetrieveFileRetries         = 3
	longDescription                    = `Profiling on existing applications with low-overhead.

These commands help you identify application performance issues.
`
	profilingExamples = `
	# Profile a pod for 5 minutes with JFR format for java language
	%[1]s prof my-pod -t 5m -l java -o jfr

	# Profile an alpine based container for java language
	%[1]s prof my-pod -l java --alpine 

	# Profile a pod for 5 minutes in intervals of 60 seconds for java language by giving the cpu limits, the container runtime, the agent image and the image pull policy
	%[1]s my-pod -l java -o flamegraph -t 5m --interval 60s --cpu-limits=1 -r containerd --image=localhost/my-agent-image-jvm:latest --image-pull-policy=IfNotPresent

	# Profile in contprof namespace a pod running in contprof-stupid-apps namespace by using the profiler service account for go language 
	%[1]s prof my-pod -n contprof --service-account=profiler --target-namespace=contprof-stupid-apps -l go

	# Set custom resource requests and limits for the agent pod (default: neither requests nor limits are set) for python language
	%[1]s prof my-pod --cpu-requests 100m --cpu-limits 200m --mem-requests 100Mi --mem-limits 200Mi -l python

	# Profile the pods with the label selector "app=my-app" for 5 minutes with JFR format for java language
	%[1]s prof -l java -o jfr -t 5m --selector app=my-app
`
)

var imagePullPolicies = []apiv1.PullPolicy{apiv1.PullNever, apiv1.PullAlways, apiv1.PullIfNotPresent}

// Profiler defines the profile method.
type Profiler interface {
	Profile(cfg *config.ProfilerConfig) error
}

type ProfileOptions struct {
	configFlags *genericclioptions.ConfigFlags
	genericiooptions.IOStreams
}

func NewProfileOptions(streams genericiooptions.IOStreams) *ProfileOptions {
	return &ProfileOptions{
		configFlags: genericclioptions.NewConfigFlags(false),
		IOStreams:   streams,
	}
}

type profilingFlags struct {
	runtime         string
	runtimePath     string
	lang            string
	event           string
	logLevel        string
	compressorType  string
	profilingTool   string
	outputType      string
	imagePullPolicy string
	privileged      bool
	capabilities    []string
}

func NewProfileCommand(streams genericiooptions.IOStreams) *cobra.Command {
	var (
		target      config.TargetConfig
		job         config.JobConfig
		showVersion bool
		flags       profilingFlags
	)

	options := NewProfileOptions(streams)
	cmd := &cobra.Command{
		Use:                   "prof [pod-name | --selector label]",
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

			if len(args) == 0 && target.LabelSelector == "" {
				_ = cmd.Help()
				return
			}

			if err := validateFlags(&flags, &target, &job); err != nil {
				_, _ = fmt.Fprintln(streams.Out, err)
				os.Exit(1)
			}

			// set log level
			level, _ := log.ParseLevel(flags.logLevel)
			log.SetLevel(level)

			if target.LabelSelector == "" {
				target.PodName = args[0]
				if len(args) > 1 {
					target.ContainerName = args[1]
				}
			}

			// Prepare profiler
			cfg, err := getProfilerConfig(target, job, flags.logLevel, flags.privileged, flags.capabilities)
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

			cfg.Job.Namespace = connectionInfo.Namespace
			err = profiler.NewJobProfiler(
				apiprof.NewPodApi(connectionInfo),
				apiprof.NewProfilingJobApi(connectionInfo),
				apiprof.NewProfilingContainerApi(connectionInfo),
			).Profile(cfg)

			if err != nil {
				printer := cli.NewPrinter(cfg.Target.DryRun)
				printer.Print("Profiling failed ... ")
				printer.PrintError()
				printer.Print("😥 " + err.Error())
			}
		},
	}

	cmd.Flags().BoolVar(&showVersion, "version", false, "Print version and build information")

	cmd.Flags().StringVar(&target.LabelSelector, "selector", "", "Label selector to target multiple pods simultaneously. Supports '=', '==', and '!=' operators (e.g. app=my-app,env=prod). All label constraints must be satisfied.")
	cmd.Flags().IntVar(&target.PoolSizeLaunchProfilingJobs, "pool-size-profiling-jobs", defaultPoolSizeLaunchProfilingJobs, "Maximum number of pods to profile in parallel when using '--selector'. Set to 0 for no limit (all matching pods profiled simultaneously)")
	cmd.Flags().StringVarP(&flags.runtime, "runtime", "r", defaultContainerRuntime,
		fmt.Sprintf("Container runtime used by the Kubernetes cluster. Choose one of: %v", api.AvailableContainerRuntimes()))
	cmd.Flags().StringVar(&flags.runtimePath, "runtime-path", api.GetContainerRuntimeRootPath[api.ContainerRuntime(defaultContainerRuntime)],
		"Root path of the container runtime installation. Override when the runtime is installed in a non-default location")

	cmd.Flags().DurationVarP(&target.Duration, "time", "t", 0, "Total profiling duration (e.g. 30s, 5m, 1h). The agent stops after this period")
	// if interval is not given, duration is set as default
	cmd.Flags().DurationVar(&target.Interval, "interval", target.Duration, "Profiling interval for continuous/iterative mode (e.g. 30s, 1m). When equal to --time a single capture is taken (discrete mode); when shorter, multiple captures are taken repeatedly until --time elapses")
	cmd.Flags().StringVar(&target.LocalPath, "local-path", "", "Local directory where result files are saved. Defaults to the current working directory")
	cmd.Flags().BoolVar(&target.Alpine, "alpine", false, "Use the Alpine-based variant of the agent image. Set this flag when the target container runs on a musl/Alpine base")
	cmd.Flags().BoolVar(&target.DryRun, "dry-run", false, "Simulate the profiling workflow without actually running the agent. Useful for validating flags and generated manifests")
	cmd.Flags().StringVar(&target.Image, "image", "", "Override the agent Docker image (e.g. my-registry/kubectl-prof-agent:latest). By default the image matching the current CLI version is used")
	cmd.Flags().StringVar(&target.Namespace, "target-namespace", "", "Kubernetes namespace of the target pod, if different from the namespace where the profiling job is created")

	cmd.Flags().StringVarP(&flags.lang, "lang", "l", "",
		fmt.Sprintf("Programming language of the target application. Choose one of: %v", api.AvailableLanguages()))
	cmd.Flags().StringVarP(&flags.event, "event", "e", defaultEvent,
		fmt.Sprintf("Profiling event to capture. Choose one of: %v", api.AvailableEvents()))

	cmd.Flags().StringVar(&job.RequestConfig.CPU, "cpu-requests", "", "CPU resource request for the agent container (e.g. 100m, 1). Defaults to no request")
	cmd.Flags().StringVar(&job.RequestConfig.Memory, "mem-requests", "", "Memory resource request for the agent container (e.g. 128Mi, 1Gi). Defaults to no request")
	cmd.Flags().StringVar(&job.LimitConfig.CPU, "cpu-limits", "", "CPU resource limit for the agent container (e.g. 200m, 2). Defaults to no limit")
	cmd.Flags().StringVar(&job.LimitConfig.Memory, "mem-limits", "", "Memory resource limit for the agent container (e.g. 256Mi, 2Gi). Defaults to no limit")
	cmd.Flags().StringVar(&target.ImagePullSecret, "image-pull-secret", "", "Name of the Kubernetes Secret of type kubernetes.io/dockerconfigjson used to pull the agent image from a private registry")
	cmd.Flags().StringVar(&target.ServiceAccountName, "service-account", "", "Name of the Kubernetes ServiceAccount to assign to the profiling agent pod")

	cmd.Flags().BoolVar(&flags.privileged, "privileged", true, "Run the agent container in privileged mode. Required by most profiling tools (perf, bpf, async-profiler, etc.)")
	cmd.Flags().StringVar(&flags.logLevel, "log-level", defaultLogLevel,
		fmt.Sprintf("Log verbosity level of the agent. Choose one of: %v", api.AvailableLogLevels()))
	/*cmd.Flags().StringVarP(&compressorType, "compressor", "c", defaultCompressor,
	fmt.Sprintf("Compressor for compressing generated profiling result, choose one of %v", compressor.AvailableCompressors()))*/
	cmd.Flags().StringVar(&flags.profilingTool, "tool", "", fmt.Sprintf("Profiling tool to use. A default tool is selected automatically based on --lang and --output when omitted. Available tools: %v", api.AvailableProfilingToolsString()))
	cmd.Flags().StringVarP(&flags.outputType, "output", "o", defaultOutputType,
		fmt.Sprintf("Output format for the profiling result. A default is selected automatically based on --tool when omitted. Available types: %v", api.AvailableOutputTypesString()))
	cmd.Flags().BoolVar(&target.PrintAgentLogs, "print-agent-logs", false, "Stream agent container logs to the local standard output while profiling is in progress")
	cmd.Flags().BoolVar(&target.PrintLogs, "print-logs", true, "Instruct the agent to forward its internal log messages to its standard output (visible via kubectl logs)")
	cmd.Flags().DurationVar(&target.GracePeriodEnding, "grace-period-ending", defaultGracePeriodEnding, "Time the CLI waits for the agent to finish and upload results before forcibly terminating the profiling job (e.g. 1m, 5m)")
	cmd.Flags().StringVar(&flags.imagePullPolicy, "image-pull-policy", defaultImagePullPolicy, fmt.Sprintf("Image pull policy for the agent container. Choose one of: %v", imagePullPolicies))
	cmd.Flags().StringVar(&target.ContainerName, "target-container-name", "", "Name of the specific container to profile inside the target pod. Required when the pod has more than one container")
	cmd.Flags().StringVar(&target.OutputSplitInChunkSize, "output-split-size", defaultOutputSplitSize, "Split large memory output files (heapdump, heapsnapshot, gcdump, dump) into chunks of this size for transfer. Uses the same format as the Unix split command (e.g. 50M, 200M, 1G)")
	cmd.Flags().IntVar(&target.PoolSizeRetrieveChunks, "pool-size-retrieve-chunks", defaultPoolSizeRetrieveChunks, "Number of parallel goroutines used to download result file chunks from the agent. Applies to chunked memory outputs (heapdump, heapsnapshot, gcdump, dump)")
	cmd.Flags().IntVar(&target.RetrieveFileRetries, "retrieve-file-retries", defaultRetrieveFileRetries, "Number of times the CLI retries downloading a result file or chunk from the agent container before failing")
	cmd.Flags().StringVar(&target.PID, "pid", "", "PID of the target process inside the container. Use when the container runs multiple processes and the main process is not PID 1")
	cmd.Flags().StringVarP(&target.Pgrep, "pgrep", "p", "", "Filter the target process by name using pgrep. Use when the PID is not known but the process name is (e.g. java, python, dotnet)")
	cmd.Flags().IntVar(&target.NodeHeapSnapshotSignal, "node-heap-snapshot-signal", 12, "OS signal number sent to the Node.js process to trigger a heap snapshot (default: 12 = SIGUSR2). Use 10 for SIGUSR1")
	cmd.Flags().StringSliceVar(&flags.capabilities, "capabilities", nil, "Linux capabilities to add to the agent container (e.g. --capabilities SYS_ADMIN --capabilities SYS_PTRACE). May be required when --privileged is false")
	cmd.Flags().StringSliceVar(&job.TolerationsRaw, "tolerations", nil, "Tolerations for the profiling job pod, in the format key=value:effect or key:effect (e.g. --tolerations node-role=infra:NoSchedule --tolerations dedicated:NoExecute)")
	cmd.Flags().StringSliceVar(&target.AsyncProfilerArgs, "async-profiler-args", nil, "Extra arguments forwarded directly to async-profiler (e.g. --async-profiler-args --alloc=2m --async-profiler-args --lock=1ms). See async-profiler docs for available options")

	options.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func getProfilerConfig(target config.TargetConfig, job config.JobConfig, logLevel string, privileged bool, capabilities []string) (*config.ProfilerConfig, error) {
	job.Privileged = privileged
	job.Capabilities = make([]apiv1.Capability, len(capabilities))
	for i, capability := range capabilities {
		job.Capabilities[i] = apiv1.Capability(capability)
	}
	return config.NewProfilerConfig(&target, config.WithJob(&job), config.WithLogLevel(api.LogLevel(logLevel)))
}

func validateFlags(flags *profilingFlags, target *config.TargetConfig, job *config.JobConfig) error {
	var err error

	err = validateLang(flags.lang)
	if err != nil {
		return err
	}

	flags.runtime, err = validateRuntime(flags.runtime, target)
	if err != nil {
		return err
	}

	if stringUtils.IsNotBlank(flags.event) && !api.IsSupportedEvent(flags.event) {
		return errors.Errorf("unsupported event, choose one of %s", api.AvailableEvents())
	}
	if stringUtils.IsBlank(flags.event) {
		flags.event = defaultEvent
	}

	if stringUtils.IsNotBlank(flags.logLevel) && !api.IsSupportedLogLevel(flags.logLevel) {
		return errors.Errorf("unsupported log level, choose one of %s", api.AvailableLogLevels())
	}
	if stringUtils.IsBlank(flags.logLevel) {
		flags.logLevel = defaultLogLevel
	}

	/*if stringUtils.IsNotBlank(compressorType) && !compressor.IsSupportedCompressor(compressorType) {
		return errors.Errorf("unsupported compressor, choose one of %s", compressor.AvailableCompressors())
	}*/
	if stringUtils.IsBlank(flags.compressorType) {
		flags.compressorType = defaultCompressor
	}

	if stringUtils.IsNotBlank(flags.imagePullPolicy) && !isSupportedImagePullPolicy(flags.imagePullPolicy) {
		return errors.Errorf("unsupported image pull policy, choose one of %s", imagePullPolicies)
	}
	if stringUtils.IsBlank(flags.imagePullPolicy) {
		flags.imagePullPolicy = defaultImagePullPolicy
	}

	target.ImagePullPolicy = apiv1.PullPolicy(flags.imagePullPolicy)
	target.Language = api.ProgrammingLanguage(flags.lang)
	target.ContainerRuntime = api.ContainerRuntime(flags.runtime)
	target.ContainerRuntimePath = api.GetContainerRuntimeRootPath[target.ContainerRuntime]
	if flags.runtimePath != api.GetContainerRuntimeRootPath[api.ContainerRuntime(defaultContainerRuntime)] {
		target.ContainerRuntimePath = flags.runtimePath
	}
	target.Event = api.ProfilingEvent(flags.event)
	target.Compressor = compressor.Type(flags.compressorType)

	validateProfilingTool(flags.profilingTool, flags.outputType, target)
	validateOutputType(flags.outputType, target)

	if _, err := job.RequestConfig.ParseResources(); err != nil {
		return errors.Wrapf(err, "unable to parse resource requests")
	}

	if _, err := job.LimitConfig.ParseResources(); err != nil {
		return errors.Wrapf(err, "unable to parse resource limits")
	}

	if err := job.ParseTolerations(); err != nil {
		return errors.Wrapf(err, "unable to parse tolerations")
	}

	// Create the local path if given and it does not exist
	if !stringUtils.IsBlank(target.LocalPath) {
		err = os.MkdirAll(target.LocalPath, 0755)
		if err != nil {
			return errors.Wrap(err, "could not create local path")
		}
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

func isSupportedImagePullPolicy(imagePullPolicy string) bool {
	for _, current := range imagePullPolicies {
		if apiv1.PullPolicy(imagePullPolicy) == current {
			return true
		}
	}
	return false
}
