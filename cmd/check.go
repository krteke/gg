package cmd

import (
	"github.com/krteke/gg/internal"
	"github.com/spf13/cobra"
)

func checkCmd() *cobra.Command {
	var job string

	cmd := &cobra.Command{
		Use:  "check <root>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := args[0]

			return internal.Check(root, job)
		},
	}

	cmd.Flags().StringVarP(&job, "job", "j", "", "job path")
	if err := cmd.MarkFlagRequired("job"); err != nil {
		panic(err)
	}

	return cmd
}
