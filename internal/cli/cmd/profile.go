package cmd

import (
	"fmt"
	config2 "github.com/josepdcs/kubectl-prof/internal/cli/config"
	kubernetes2 "github.com/josepdcs/kubectl-prof/internal/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/internal/cli/profiler"
	"github.com/josepdcs/kubectl-prof/internal/cli/version"
	log "github.com/sirupsen/logrus"
	"os"
	"time"

	"github.com/josepdcs/kubectl-prof/api"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	defaultDuration   = 1 * time.Minute
	defaultEvent      = string(api.Cpu)
	defaultLogLevel   = string(api.InfoLevel)
	defaultCompressor = string(api.Snappy)
	defaultOutputType = string(api.FlameGraph)
	longDescription   = `Profiling on existing applications with low-overhead.

These commands help you identify application performance issues.
`
	profilingExamples = `
	# Profile a pod for 5 minutes and save the output as flame.html file for java language
	%[1]s prof mypod -f flame.html -t 5m -l java

	# Profile an alpine based container for java language
	%[1]s prof mypod -f flame.html -l java --alpine 

	# Profile a pod for 5 minutes and save the output as flight.jfr file with JFR format for java language  
	%[1]s prof mypod -f flight.jfr -t 5m -l java -o jfr

	# Profile specific container container1 from pod mypod in namespace test for go language
	%[1]s prof mypod -f /tmp/flame.svg -n test container1 -l go

	# Set custom resource requests and limits for the agent pod (default: neither requests nor limits are set) for python language
	%[1]s prof mypod -f flame.svg -cpu.requests 100m -cpu.limits 200m -mem.requests 100Mi -mem.limits 200Mi -l python
`
)

type Profiler interface {
	Profile(cfg *config2.ProfilerConfig)
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
		target        config2.TargetConfig
		job           config2.JobConfig
		showVersion   bool
		runtime       string
		lang          string
		event         string
		logLevel      string
		compressor    string
		profilingTool string
		outputType    string
	)

	options := NewProfileOptions(streams)
	cmd := &cobra.Command{
		Use:                   "prof [pod-name]",
		DisableFlagsInUseLine: true,
		Short:                 "Profile running applications. Several output types are supported: flamegraphs, jfrs, threadump, heapdumps, etc.",
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

			if err := validateFlags(runtime, lang, event, logLevel, compressor, profilingTool, outputType, &target, &job); err != nil {
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
			cfg := config2.NewProfilerConfig(&target, &job).WithLogLevel(api.LogLevel(logLevel))

			connector := kubernetes2.NewConnector()
			connectionContext, err := connector.Connect(options.configFlags)
			if err != nil {
				log.Fatalf("Failed connecting to kubernetes cluster: %v\n", err)
			}

			if cfg.Target.Namespace == "" {
				cfg.Target.Namespace = connectionContext.Namespace
			}
			cfg.Job.Namespace = connectionContext.Namespace

			getter := kubernetes2.NewGetter(connectionContext.KubeContext)
			creator := kubernetes2.NewCreator(connectionContext.ClientSet)
			deleter := kubernetes2.NewDeleter(connectionContext.ClientSet)
			profiler.NewProfiler(getter, creator, deleter).Profile(cfg)
		},
	}

	cmd.Flags().BoolVar(&showVersion, "version", false, "Print version info")

	cmd.Flags().StringVarP(&runtime, "runtime", "r", "crio",
		fmt.Sprintf("The container runtime used for kubernetes, choose one of %v", api.AvailableContainerRuntimes()))
	cmd.Flags().StringVar(&target.ContainerRuntimePath, "runtime-path", api.GetContainerRuntimeRootPath[api.Crio],
		"Use a different container runtime install path")

	cmd.Flags().DurationVarP(&target.Duration, "time", "t", defaultDuration, "Max scan Duration")
	cmd.Flags().StringVarP(&target.FileName, "file", "f", "flamegraph.svg", "Optional file location")
	cmd.Flags().BoolVar(&target.Alpine, "alpine", false, "TargetConfig image is based on Alpine")
	cmd.Flags().BoolVar(&target.DryRun, "dry-run", false, "Simulate profiling")
	cmd.Flags().StringVar(&target.Image, "image", "", "Manually choose agent docker image")
	cmd.Flags().StringVar(&target.Namespace, "target-namespace", "", "namespace of target pod if different from job namespace")
	cmd.Flags().StringVarP(&target.Pgrep, "pgrep", "p", "", "name of the target process")

	cmd.Flags().StringVarP(&lang, "lang", "l", "",
		fmt.Sprintf("Programming language of the target application, choose one of %v", api.AvailableLanguages()))
	cmd.Flags().StringVarP(&event, "event", "e", defaultEvent,
		fmt.Sprintf("Profiling event, choose one of %v", api.AvailableEvents()))

	cmd.Flags().StringVar(&job.RequestConfig.CPU, "cpu.requests", "", "CPU requests of the started profiling container")
	cmd.Flags().StringVar(&job.RequestConfig.Memory, "mem.requests", "", "Memory requests of the started profiling container")
	cmd.Flags().StringVar(&job.LimitConfig.CPU, "cpu.limits", "", "CPU limits of the started profiling container")
	cmd.Flags().StringVar(&job.LimitConfig.Memory, "mem.limits", "", "Memory limits of the started profiling container")
	cmd.Flags().StringVar(&target.ImagePullSecret, "imagePullSecret", "", "imagePullSecret for agent docker image")
	cmd.Flags().StringVar(&target.ServiceAccountName, "serviceAccountName", "", "serviceAccountName to be used for profiling container")

	cmd.Flags().BoolVar(&job.Privileged, "privileged", false, "Run agent container in privileged mode")
	cmd.Flags().StringVar(&logLevel, "log-level", defaultLogLevel,
		fmt.Sprintf("Log level, choose one of %v", api.AvailableLogLevels()))
	cmd.Flags().StringVarP(&compressor, "compressor", "c", defaultCompressor,
		fmt.Sprintf("Compressor for compressing generated profiling result, choose one of %v", api.AvailableCompressors()))
	cmd.Flags().StringVar(&profilingTool, "tool", "", fmt.Sprintf("Profiling tool, choose one accorfing language %v", api.AvailableProfilingToolsString()))
	cmd.Flags().StringVarP(&outputType, "output", "o", defaultOutputType,
		fmt.Sprintf("Output type, choose one accorting tool %v", api.AvailableOutputTypesString()))

	options.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func validateFlags(runtime string, lang string, event string, logLevel string, compressor string, profilingTool string,
	outputType string, target *config2.TargetConfig, job *config2.JobConfig) error {
	if lang == "" {
		return fmt.Errorf("use -l flag to select one of the supported languages %s", api.AvailableLanguages())
	}

	if !api.IsSupportedLanguage(lang) {
		return fmt.Errorf("unsupported language, choose one of %s", api.AvailableLanguages())
	}

	if runtime != "" && !api.IsSupportedContainerRuntime(runtime) {
		return fmt.Errorf("unsupported container runtime, choose one of %s", api.AvailableContainerRuntimes())
	}

	if event != "" && !api.IsSupportedEvent(event) {
		return fmt.Errorf("unsupported event, choose one of %s", api.AvailableEvents())
	}

	if logLevel != "" && !api.IsSupportedLogLevel(logLevel) {
		return fmt.Errorf("unsupported log level, choose one of %s", api.AvailableLogLevels())
	}

	if compressor != "" && !api.IsSupportedCompressor(compressor) {
		return fmt.Errorf("unsupported compressor, choose one of %s", api.AvailableCompressors())
	}

	target.Language = api.ProgrammingLanguage(lang)
	target.ContainerRuntime = api.ContainerRuntime(runtime)
	target.ContainerRuntimePath = api.GetContainerRuntimeRootPath[target.ContainerRuntime]
	target.Event = api.ProfilingEvent(event)
	target.Compressor = api.Compressor(compressor)

	validateProfilingTool(profilingTool, target)
	validateOutputType(outputType, target)

	if _, err := job.RequestConfig.ParseResources(); err != nil {
		return fmt.Errorf("unable to parse resource requests: %w", err)
	}

	if _, err := job.LimitConfig.ParseResources(); err != nil {
		return fmt.Errorf("unable to parse resourse limits: %w", err)
	}

	return nil
}

func validateProfilingTool(profilingTool string, target *config2.TargetConfig) {
	defaultTool := api.GetProfilingToolsByProgrammingLanguage[target.Language][0]
	if profilingTool == "" {
		fmt.Printf("Default profiling tool %s will be used ... ✔\n", defaultTool)
		target.ProfilingTool = defaultTool
		return
	}

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

func validateOutputType(outputType string, target *config2.TargetConfig) {
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

	if !api.IsValidOutputType(api.EventType(outputType), target.ProfilingTool) {
		fmt.Printf("Unsupported output type %s for profiling tool %s, default %s will be used ... ✔\n",
			outputType, target.ProfilingTool, defaultOutputType)
		target.OutputType = defaultOutputType
		return
	}

	target.OutputType = api.EventType(outputType)
}
