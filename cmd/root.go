package cmd

import (
	"github.com/krteke/gg/internal"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	var verbose int
	// var quiet bool

	cmd := &cobra.Command{
		Use:   "gg",
		Short: "cli",
	}

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	cmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "")
	// cmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "")

	internal.InitLogger(verbose)

	cmd.AddCommand(scanCmd(), checkCmd())

	return cmd
}

func Execute() error {
	return rootCmd().Execute()
}
