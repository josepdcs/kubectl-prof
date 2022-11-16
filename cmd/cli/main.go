package main

import (
	"github.com/josepdcs/kubectl-prof/internal/cli/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)

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
