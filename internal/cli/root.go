package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/HobaiRiku/hosta/internal/app"
	"github.com/HobaiRiku/hosta/internal/launcher"
	"github.com/HobaiRiku/hosta/internal/openssh"
	"github.com/HobaiRiku/hosta/internal/sshconfig"
	"github.com/HobaiRiku/hosta/internal/version"
	"github.com/spf13/cobra"
)

type sshClient interface {
	Binary() string
	Version(context.Context) (string, error)
	Resolve(context.Context, string, string, bool) (openssh.Resolved, error)
	Connect(context.Context, string, string, bool) error
}

type interactiveSSHClient interface {
	ConnectInteractive(context.Context, string, string, bool) error
}

type dependencies struct {
	load   func(string) (*app.Snapshot, error)
	newSSH func() (sshClient, error)
}

type rootOptions struct {
	configPath  string
	showVersion bool
}

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	home, _ := os.UserHomeDir()
	defaultPath := filepath.Join(home, ".ssh", "config")
	deps := dependencies{
		load: func(path string) (*app.Snapshot, error) {
			return app.Load(path, sshconfig.Options{})
		},
		newSSH: func() (sshClient, error) {
			return openssh.New("", nil)
		},
	}
	return newRootCommand(deps, defaultPath)
}

func newRootCommand(deps dependencies, defaultConfigPath string) *cobra.Command {
	options := &rootOptions{}
	root := &cobra.Command{
		Use:           "hosta",
		Short:         "A fast, interactive SSH host launcher",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if options.showVersion {
				return writeVersion(cmd.OutOrStdout())
			}
			snapshot, err := loadSnapshot(deps, options.configPath)
			if err != nil {
				return err
			}
			commandForAlias := func(alias string) (string, error) {
				client, err := deps.newSSH()
				if err != nil {
					return "", err
				}
				resolved, err := client.Resolve(cmd.Context(), alias, options.configPath, configWasExplicit(cmd))
				if err != nil {
					return "", err
				}
				return resolved.ConnectionCommand()
			}
			alias, err := launcher.Run(cmd.InOrStdin(), cmd.OutOrStdout(), snapshot.Index, commandForAlias)
			if err != nil {
				return err
			}
			if alias == "" {
				return nil
			}
			client, err := deps.newSSH()
			if err != nil {
				return withCode(4, err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Connecting to %s ...\n", alias); err != nil {
				return err
			}
			if interactive, ok := client.(interactiveSSHClient); ok {
				return interactive.ConnectInteractive(cmd.Context(), alias, options.configPath, configWasExplicit(cmd))
			}
			return client.Connect(cmd.Context(), alias, options.configPath, configWasExplicit(cmd))
		},
	}
	root.PersistentFlags().StringVar(&options.configPath, "config", defaultConfigPath, "path to the user SSH config")
	root.Flags().BoolVarP(&options.showVersion, "version", "v", false, "print version and exit")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return withCode(2, err) })
	root.AddCommand(
		newVersionCommand(),
		newListCommand(deps, options),
		newShowCommand(deps, options),
		newDoctorCommand(deps, options),
		newConnectCommand(deps, options),
		newConfigCommand(options),
		newCompletionCommand(),
	)
	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeVersion(cmd.OutOrStdout())
		},
	}
}

func writeVersion(w io.Writer) error {
	_, err := fmt.Fprintln(w, version.String())
	return err
}
