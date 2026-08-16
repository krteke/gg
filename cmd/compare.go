package cmd

import (
	"github.com/krteke/gg/internal"
	"github.com/spf13/cobra"
)

func compareCmd() *cobra.Command {
	var config internal.CompareConfig

	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.Compare()
		},
	}

	cmd.Flags().StringVarP(&config.Job, "job", "j", "", "")
	cmd.Flags().StringVarP(&config.Source, "source", "s", "", "")
	cmd.Flags().StringVarP(&config.Target, "target", "t", "", "")
	cmd.Flags().StringVarP(&config.Output, "output", "o", "compare", "")

	cmd.MarkFlagsRequiredTogether("job", "source", "target")

	return cmd
}
