package main

import (
	"github.com/josepdcs/kubectl-profile/internal/cli"
	"github.com/spf13/cobra"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	flags := pflag.NewFlagSet("cli", pflag.ExitOnError)
	pflag.CommandLine = flags

	streams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	root := cli.NewFlameCommand(streams)
	cobra.CheckErr(root.Execute())
}
