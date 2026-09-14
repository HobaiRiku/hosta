package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConnectCommand(deps dependencies, options *rootOptions) *cobra.Command {
	command := &cobra.Command{
		Use:     "connect <host>",
		Aliases: []string{"c"},
		Short:   "Connect to a discovered SSH host",
		Args:    exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshot, err := loadSnapshot(deps, options.configPath)
			if err != nil {
				return err
			}
			value, ok := snapshot.Index.Get(args[0])
			if !ok {
				return withCode(2, fmt.Errorf("SSH host %q was not discovered", args[0]))
			}
			client, err := deps.newSSH()
			if err != nil {
				return withCode(4, err)
			}
			return client.Connect(cmd.Context(), value.Alias, options.configPath, configWasExplicit(cmd))
		},
	}
	command.ValidArgsFunction = completeHosts(deps, options)
	return command
}
