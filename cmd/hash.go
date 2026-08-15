package cmd

import (
	"github.com/krteke/gg/internal"
	"github.com/spf13/cobra"
)

func hashCmd() *cobra.Command {
	var job string
	var output string

	cmd := &cobra.Command{
		Use:  "hash <root>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := args[0]

			return internal.Hash(root)
		},
	}

	cmd.Flags().StringVarP(&job, "job", "j", "", "")
	cmd.Flags().StringVarP(&output, "output", "o", "src", "")

	if err := cmd.MarkFlagRequired("job"); err != nil {
		panic(err)
	}

	return cmd
}
