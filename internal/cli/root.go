package cli

import (
	"fmt"
	"io"

	"github.com/HobaiRiku/hosta/internal/buildinfo"
	"github.com/HobaiRiku/hosta/internal/launcher"
	"github.com/spf13/cobra"
)

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	info := buildinfo.Current()
	root := &cobra.Command{
		Use:           "hosta",
		Short:         "A fast, interactive SSH host launcher",
		Version:       info.Version,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return launcher.Run(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	root.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	root.AddCommand(newVersionCommand(info))
	return root
}

func newVersionCommand(info buildinfo.Info) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeVersion(cmd.OutOrStdout(), info)
		},
	}
}

func writeVersion(w io.Writer, info buildinfo.Info) error {
	_, err := fmt.Fprintf(w, "hosta %s\ncommit: %s\nbuilt: %s\n", info.Version, info.Commit, info.Date)
	return err
}
