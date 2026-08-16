package cmd

import (
	"github.com/krteke/gg/internal"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	var verbose int

	cmd := &cobra.Command{
		Use:   "gg",
		Short: "cli",
	}

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	cmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "")

	internal.InitLogger(verbose)

	cmd.AddCommand(scanCmd(), checkCmd(), hashCmd(), compareCmd())

	return cmd
}

func Execute() error {
	return rootCmd().Execute()
}
