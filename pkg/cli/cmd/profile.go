package cmd

import (
	"fmt"
	"github.com/josepdcs/kubectl-profile/pkg/cli/config"
	"github.com/josepdcs/kubectl-profile/pkg/cli/profiler"
	"github.com/josepdcs/kubectl-profile/pkg/cli/version"
	"os"
	"time"

	"github.com/josepdcs/kubectl-profile/api"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	defaultDuration = 1 * time.Minute
	defaultEvent    = string(api.Cpu)
	flameLong       = `Profiling on existing applications with low-overhead by generating flame graphs.

These commands help you identify application performance issues.
`
	flameExamples = `
	# ProfileConfig a pod for 5 minutes and save the output as flame.svg file
	%[1]s profile mypod -f flame.svg -t 5m

	# ProfileConfig an alpine based container
	%[1]s profile mypod -f flame.svg --alpine

	# ProfileConfig specific container container1 from pod mypod in namespace test
	%[1]s profile mypod -f /tmp/flame.svg -n test container1

	# Set custom resource requests and limits for the cli pod (default: neither requests nor limits are set)
	%[1]s profile mypod -f flame.svg -cpu.requests 100m -cpu.limits 200m -mem.requests 100Mi -mem.limits 200Mi
`
)

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
		target        config.TargetConfig
		job           config.JobConfig
		showVersion   bool
		chosenRuntime string
		chosenLang    string
		chosenEvent   string
	)

	options := NewProfileOptions(streams)
	cmd := &cobra.Command{
		Use:                   "profile [pod-name]",
		DisableFlagsInUseLine: true,
		Short:                 "ProfileConfig running applications by generating flame graphs.",
		Long:                  flameLong,
		Example:               fmt.Sprintf(flameExamples, "kubectl"),
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

			if err := validateFlags(chosenRuntime, chosenLang, chosenEvent, &target, &job); err != nil {
				_, _ = fmt.Fprintln(streams.Out, err)
				os.Exit(1)
			}

			target.PodName = args[0]
			if len(args) > 1 {
				target.ContainerName = args[1]
			}

			cfg := &config.ProfileConfig{
				Target:      &target,
				Job:         &job,
				ConfigFlags: options.configFlags,
			}

			profiler.Profile(cfg)
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
	cmd.Flags().StringVar(&target.Namespace, "target-namespace", "", "namespace of target pod if differnt from job namespace")
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

	options.configFlags.AddFlags(cmd.Flags())

	return cmd
}

func validateFlags(runtimeString string, langString string, eventString string, targetDetails *config.TargetConfig, jobDetails *config.JobConfig) error {
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

	targetDetails.Language = api.ProgrammingLanguage(langString)
	targetDetails.ContainerRuntime = api.ContainerRuntime(runtimeString)
	targetDetails.Event = api.ProfilingEvent(eventString)

	if _, err := jobDetails.RequestConfig.ParseResources(); err != nil {
		return fmt.Errorf("unable to parse resource requests: %w", err)
	}

	if _, err := jobDetails.LimitConfig.ParseResources(); err != nil {
		return fmt.Errorf("unable to parse resourse limits: %w", err)
	}

	return nil
}
