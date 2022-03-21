package main

import (
	"github.com/josepdcs/kubectl-prof/pkg/cli/cmd"
	"github.com/spf13/cobra"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-prof", pflag.ExitOnError)
	pflag.CommandLine = flags

	streams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	profileCmd := cmd.NewProfileCommand(streams)
	cobra.CheckErr(profileCmd.Execute())
}
