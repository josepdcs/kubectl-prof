package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/josepdcs/kubectl-prof/internal/cli"
	"github.com/josepdcs/kubectl-prof/internal/cli/config"
	"github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/profiler"
	apiprof "github.com/josepdcs/kubectl-prof/internal/cli/profiler/api"
	"github.com/josepdcs/kubectl-prof/internal/cli/version"
	"github.com/josepdcs/kubectl-prof/pkg/util/compressor"
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

// imagePullPolicies defines a list of container image pull policies supported by the Kubernetes API.
var imagePullPolicies = []apiv1.PullPolicy{apiv1.PullNever, apiv1.PullAlways, apiv1.PullIfNotPresent}

// Profiler defines the profile method.
type Profiler interface {
	Profile(cfg *config.ProfilerConfig) error
}

// ProfileOptions holds configuration flags and IO streams for profile-related commands.
type ProfileOptions struct {
	configFlags *genericclioptions.ConfigFlags
	genericiooptions.IOStreams
}

// NewProfileOptions initializes and returns a new instance of ProfileOptions with the provided IOStreams.
func NewProfileOptions(streams genericiooptions.IOStreams) *ProfileOptions {
	return &ProfileOptions{
		configFlags: genericclioptions.NewConfigFlags(false),
		IOStreams:   streams,
	}
}

// profilingFlags represents configurable options for profiling operations, such as runtime, language, output type, and more.
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

// profilingContext contains the necessary context to execute the profiling command.
type profilingContext struct {
	cmd         *cobra.Command
	args        []string
	streams     genericiooptions.IOStreams
	target      *config.TargetConfig
	job         *config.JobConfig
	flags       *profilingFlags
	showVersion bool
	options     *ProfileOptions
}

// NewProfile returns a new cobra.Command for the "prof" subcommand.
// This command allows users to profile running applications in a Kubernetes cluster.
// It supports multiple output types such as flamegraphs, JFRs, thread dumps, and heap dumps.
func NewProfile(streams genericiooptions.IOStreams) *cobra.Command {
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
			ctx := &profilingContext{
				cmd:         cmd,
				args:        args,
				streams:     streams,
				target:      &target,
				job:         &job,
				flags:       &flags,
				showVersion: showVersion,
				options:     options,
			}
			runProfile(ctx)
		},
	}

	setProfileFlags(cmd, &target, &job, &flags, &showVersion, options)

	return cmd
}

// runProfile executes the profiling logic when the "prof" command is invoked.
func runProfile(ctx *profilingContext) {
	if ctx.showVersion {
		_, _ = fmt.Fprintln(ctx.streams.Out, version.String())
		return
	}

	if len(ctx.args) == 0 && ctx.target.LabelSelector == "" {
		_ = ctx.cmd.Help()
		return
	}

	if err := validateFlags(ctx.flags, ctx.target, ctx.job); err != nil {
		_, _ = fmt.Fprintln(ctx.streams.Out, err)
		os.Exit(1)
	}

	// set log level
	level, _ := log.ParseLevel(ctx.flags.logLevel)
	log.SetLevel(level)

	if ctx.target.LabelSelector == "" {
		ctx.target.PodName = ctx.args[0]
		if len(ctx.args) > 1 {
			ctx.target.ContainerName = ctx.args[1]
		}
	}

	// Prepare profiler
	cfg, err := getProfilerConfig(*ctx.target, *ctx.job, ctx.flags.logLevel, ctx.flags.privileged, ctx.flags.capabilities)
	if err != nil {
		log.Fatalf("Failed configure profiler: %v\n", err)
	}

	connectionInfo, err := kubernetes.Connect(ctx.options.configFlags)
	if err != nil {
		log.Fatalf("Failed connecting to kubernetes cluster: %v\n", err)
	}

	if cfg.Target.Namespace == "" {
		cfg.Target.Namespace = connectionInfo.Namespace
	}

	cfg.Job.Namespace = connectionInfo.Namespace
	err = profiler.New(
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
}

// setProfileFlags defines and binds all CLI flags for the "prof" command.
func setProfileFlags(cmd *cobra.Command, target *config.TargetConfig, job *config.JobConfig, flags *profilingFlags, showVersion *bool, options *ProfileOptions) {
	cmd.Flags().BoolVar(showVersion, "version", false, "Print version and build information")

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
	cmd.Flags().DurationVar(&target.HeartbeatInterval, "heartbeat-interval", 30*time.Second, "Interval between heartbeat progress events emitted during profiling. Keeps connections alive through proxies/load balancers (e.g. 30s, 1m)")
	cmd.Flags().StringSliceVar(&target.AsyncProfilerArgs, "async-profiler-args", nil, "Extra arguments forwarded directly to async-profiler (e.g. --async-profiler-args --alloc=2m --async-profiler-args --lock=1ms). See async-profiler docs for available options")
	cmd.Flags().StringVar(&target.PprofPort, "pprof-port", "", "Port on the target pod where the Go pprof HTTP endpoint is exposed (default: 6060). Used only with --tool pprof")

	options.configFlags.AddFlags(cmd.Flags())
}

// getProfilerConfig creates a config.ProfilerConfig based on the provided target and job configurations,
// log level, privileged status, and Linux capabilities.
func getProfilerConfig(target config.TargetConfig, job config.JobConfig, logLevel string, privileged bool, capabilities []string) (*config.ProfilerConfig, error) {
	job.Privileged = privileged
	job.Capabilities = make([]apiv1.Capability, len(capabilities))
	for i, capability := range capabilities {
		job.Capabilities[i] = apiv1.Capability(capability)
	}
	return config.NewProfilerConfig(&target, config.WithJob(&job), config.WithLogLevel(api.LogLevel(logLevel)))
}
