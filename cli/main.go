package main

import (
	"github.com/spf13/cobra"
	"os"

	"github.com/josepdcs/kubectl-flame/cli/cmd"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-flame", pflag.ExitOnError)
	pflag.CommandLine = flags

	streams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	root := cmd.NewFlameCommand(streams)
	cobra.CheckErr(root.Execute())
}
