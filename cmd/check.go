package cmd

import (
	"github.com/krteke/gg/internal"
	"github.com/spf13/cobra"
)

func checkCmd() *cobra.Command {
	var config internal.CheckConfig

	cmd := &cobra.Command{
		Use:  "check <root>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := args[0]
			config.Root = root

			return config.Check()
		},
	}

	cmd.Flags().StringVarP(&config.Job, "job", "j", "", "job path")
	cmd.Flags().StringVarP(&config.Report, "report", "r", "", "")
	if err := cmd.MarkFlagRequired("job"); err != nil {
		panic(err)
	}

	return cmd
}
