package cmd

import (
	"github.com/krteke/gg/internal"
	"github.com/spf13/cobra"
)

func compareCmd() *cobra.Command {
	var config internal.CompareConfig

	cmd := &cobra.Command{
		Use:  "compare",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.Compare()
		},
	}

	cmd.Flags().StringVarP(&config.Job, "job", "j", "", "job directory")
	cmd.Flags().StringVarP(&config.Source, "source", "s", "", "source hash directory")
	cmd.Flags().StringVarP(&config.Target, "target", "t", "", "target hash directory")
	cmd.Flags().StringVarP(&config.Output, "output", "o", "compare", "output directory")

	for _, flag := range []string{"job", "source", "target"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}

	return cmd
}
