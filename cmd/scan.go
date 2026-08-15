package cmd

import (
	"github.com/krteke/gg/internal"
	"github.com/spf13/cobra"
)

func scanCmd() *cobra.Command {
	var config internal.ScanConfig

	cmd := &cobra.Command{
		Use:  "scan <root>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := args[0]
			config := internal.ScanConfig{
				Root:   root,
				Output: config.Output,
				Max:    config.Max,
			}

			return internal.Scan(config)
		},
	}

	cmd.Flags().StringVarP(&config.Output, "output", "o", "", "output directory")
	cmd.Flags().StringVarP(&config.Max, "max", "m", "400GiB", "max bucket size")

	return cmd
}
