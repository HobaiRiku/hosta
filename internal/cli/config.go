package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print the SSH config entry path",
		Args:  exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), options.configPath)
			return err
		},
	}
}

func configWasExplicit(cmd *cobra.Command) bool {
	return cmd.Flags().Changed("config")
}
