package cmd

import (
	"github.com/krteke/gg/internal"
	"github.com/spf13/cobra"
)

func hashCmd() *cobra.Command {
	var config internal.HashConfig
	var retry int

	cmd := &cobra.Command{
		Use:  "hash <root>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := args[0]
			config.Root = root
			if cmd.Flags().Changed("retry") {
				config.Retry = &retry
			}

			return config.Hash()
		},
	}

	cmd.Flags().StringVarP(&config.Job, "job", "j", "", "")
	cmd.Flags().StringVarP(&config.Output, "output", "o", "src", "")
	cmd.Flags().Uint32Var(&config.At, "at", 0, "start at which bucket")
	cmd.Flags().IntVarP(&retry, "retry", "r", 0, "retry which bucket")

	if err := cmd.MarkFlagRequired("job"); err != nil {
		panic(err)
	}

	return cmd
}
