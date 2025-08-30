package cmd

import (
	"os"

	"github.com/apex/log"
	"github.com/spf13/cobra"
)

var (
	configFile string
)

type rootCmd struct {
	cmd   *cobra.Command
	debug bool
}

func Execute(args []string) {
	cmd := newRootCmd()
	if err := cmd.Execute(args); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *rootCmd {
	root := &rootCmd{}
	cmd := &cobra.Command{
		Use:               "dockerfiles",
		Short:             "Generate and manage Docker base images from templates",
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		PersistentPreRun: func(*cobra.Command, []string) {
			if root.debug {
				log.SetLevel(log.DebugLevel)
				log.Debug("verbose output enabled")
			}
		},
		PersistentPostRun: func(*cobra.Command, []string) {
			log.Info("thanks for using Dockerfiles!")
		},
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Load configuration from file")
	_ = cmd.MarkFlagFilename("config", "yaml", "yml")
	cmd.PersistentFlags().BoolVar(&root.debug, "debug", false, "Enable debug logging and verbose output")

	cmd.AddCommand(
		newGeneratorCmd().Cmd,
		newCleanCmd().Cmd,
	)
	root.cmd = cmd
	return root
}

func (cmd *rootCmd) Execute(args []string) error {
	cmd.cmd.SetArgs(args)

	if err := cmd.cmd.Execute(); err != nil {
		log.WithError(err).Error("command failed")
		return err
	}

	return nil
}
