package cmd

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/pkg/cli/config"
	"github.com/josepdcs/kubectl-prof/pkg/cli/kubernetes"
	"github.com/josepdcs/kubectl-prof/pkg/cli/profiler"
	"github.com/josepdcs/kubectl-prof/pkg/cli/version"
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
	longDescription   = `Profiling on existing applications with low-overhead.

These commands help you identify application performance issues.
`
	profilingExamples = `
	# Profile a pod for 5 minutes and save the output as flame.html file for java language
	%[1]s prof mypod -f flame.html -t 5m -l java

	# Profile an alpine based container for java language
	%[1]s prof mypod -f flame.html -l java --alpine 

	# Profile specific container container1 from pod mypod in namespace test for go language
	%[1]s prof mypod -f /tmp/flame.svg -n test container1 -l go

	# Set custom resource requests and limits for the cli pod (default: neither requests nor limits are set) for python language
	%[1]s prof mypod -f flame.svg -cpu.requests 100m -cpu.limits 200m -mem.requests 100Mi -mem.limits 200Mi -l python
`
)

type Profiler interface {
	Profile(cfg *config.ProfilerConfig)
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
		target         config.TargetConfig
		job            config.JobConfig
		showVersion    bool
		chosenRuntime  string
		chosenLang     string
		chosenEvent    string
		chosenLogLevel string
		compressor     string
	)

	options := NewProfileOptions(streams)
	cmd := &cobra.Command{
		Use:                   "prof [pod-name]",
		DisableFlagsInUseLine: true,
		Short:                 "Profile running applications by generating flame graphs at the moment.",
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

			if err := validateFlags(chosenRuntime, chosenLang, chosenEvent, chosenLogLevel, compressor, &target, &job); err != nil {
				_, _ = fmt.Fprintln(streams.Out, err)
				os.Exit(1)
			}

			// set log level
			level, _ := log.ParseLevel(chosenLogLevel)
			log.SetLevel(level)

			target.PodName = args[0]
			if len(args) > 1 {
				target.ContainerName = args[1]
			}

			// Prepare profiler
			cfg := config.NewProfilerConfig(&target, &job, options.configFlags).
				WithLogLevel(api.LogLevel(chosenLogLevel))

			connector := kubernetes.NewConnector()
			getter := kubernetes.NewGetter()
			creator := kubernetes.NewCreator()
			deleter := kubernetes.NewDeleter()
			profiler.NewProfiler(connector, getter, creator, deleter).Profile(cfg)
		},
	}

	cmd.Flags().BoolVar(&showVersion, "version", false, "Print version info")

	cmd.Flags().StringVarP(&chosenRuntime, "runtime", "r", "crio",
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

	cmd.Flags().StringVarP(&chosenLang, "lang", "l", "",
		fmt.Sprintf("Programming language of the target application, choose one of %v", api.AvailableLanguages()))
	cmd.Flags().StringVarP(&chosenEvent, "event", "e", defaultEvent,
		fmt.Sprintf("Profiling event, choose one of %v", api.AvailableEvents()))

	cmd.Flags().StringVar(&job.RequestConfig.CPU, "cpu.requests", "", "CPU requests of the started profiling container")
	cmd.Flags().StringVar(&job.RequestConfig.Memory, "mem.requests", "", "Memory requests of the started profiling container")
	cmd.Flags().StringVar(&job.LimitConfig.CPU, "cpu.limits", "", "CPU limits of the started profiling container")
	cmd.Flags().StringVar(&job.LimitConfig.Memory, "mem.limits", "", "Memory limits of the started profiling container")
	cmd.Flags().StringVar(&target.ImagePullSecret, "imagePullSecret", "", "imagePullSecret for agent docker image")
	cmd.Flags().StringVar(&target.ServiceAccountName, "serviceAccountName", "", "serviceAccountName to be used for profiling container")

	cmd.Flags().BoolVar(&job.Privileged, "privileged", false, "run agent container in privileged mode")
	cmd.Flags().StringVar(&chosenLogLevel, "log-level", defaultLogLevel,
		fmt.Sprintf("Log level, choose one of %v", api.AvailableLogLevels()))
	cmd.Flags().StringVarP(&compressor, "compressor", "c", defaultCompressor,
		fmt.Sprintf("Compressor for compressing generated profiling result, choose one of %v", api.AvailableCompressors()))

	options.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func validateFlags(runtimeString string, langString string, eventString string, logLevelString string, compressorString string,
	target *config.TargetConfig, job *config.JobConfig) error {
	if langString == "" {
		return fmt.Errorf("use -l flag to select one of the supported languages %s", api.AvailableLanguages())
	}

	if !api.IsSupportedLanguage(langString) {
		return fmt.Errorf("unsupported language, choose one of %s", api.AvailableLanguages())
	}

	if runtimeString != "" && !api.IsSupportedContainerRuntime(runtimeString) {
		return fmt.Errorf("unsupported container runtime, choose one of %s", api.AvailableContainerRuntimes())
	}

	if eventString != "" && !api.IsSupportedEvent(eventString) {
		return fmt.Errorf("unsupported event, choose one of %s", api.AvailableEvents())
	}

	if logLevelString != "" && !api.IsSupportedLogLevel(logLevelString) {
		return fmt.Errorf("unsupported log level, choose one of %s", api.AvailableLogLevels())
	}

	if compressorString != "" && !api.IsSupportedCompressor(compressorString) {
		return fmt.Errorf("unsupported compressor, choose one of %s", api.AvailableCompressors())
	}

	target.Language = api.ProgrammingLanguage(langString)
	target.ContainerRuntime = api.ContainerRuntime(runtimeString)
	target.ContainerRuntimePath = api.GetContainerRuntimeRootPath[target.ContainerRuntime]
	target.Event = api.ProfilingEvent(eventString)
	target.Compressor = api.Compressor(compressorString)

	if _, err := job.RequestConfig.ParseResources(); err != nil {
		return fmt.Errorf("unable to parse resource requests: %w", err)
	}

	if _, err := job.LimitConfig.ParseResources(); err != nil {
		return fmt.Errorf("unable to parse resourse limits: %w", err)
	}

	return nil
}
