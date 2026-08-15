package cmd

import (
	"time"

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
				Config: config.Config,
			}

			return internal.Scan(config)
		},
	}

	defaultOutput := "job-" + time.Now().Format("20260102")
	cmd.Flags().StringVarP(&config.Output, "output", "o", defaultOutput, "output directory")
	cmd.Flags().StringVarP(&config.Config, "config", "c", "", "path of config file")

	return cmd
}
