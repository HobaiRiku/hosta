package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/HobaiRiku/hosta/internal/app"
	"github.com/HobaiRiku/hosta/internal/buildinfo"
	"github.com/HobaiRiku/hosta/internal/launcher"
	"github.com/HobaiRiku/hosta/internal/openssh"
	"github.com/HobaiRiku/hosta/internal/sshconfig"
	"github.com/spf13/cobra"
)

type sshClient interface {
	Binary() string
	Version(context.Context) (string, error)
	Resolve(context.Context, string, string, bool) (openssh.Resolved, error)
	Connect(context.Context, string, string, bool) error
}

type dependencies struct {
	load   func(string) (*app.Snapshot, error)
	newSSH func() (sshClient, error)
}

type rootOptions struct {
	configPath string
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
	info := buildinfo.Current()
	options := &rootOptions{}
	root := &cobra.Command{
		Use:           "hosta",
		Short:         "A fast, interactive SSH host launcher",
		Version:       info.Version,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return launcher.Run(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	root.PersistentFlags().StringVar(&options.configPath, "config", defaultConfigPath, "path to the user SSH config")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return withCode(2, err) })
	root.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	root.AddCommand(
		newVersionCommand(info),
		newListCommand(deps, options),
		newShowCommand(deps, options),
		newDoctorCommand(deps, options),
		newConnectCommand(deps, options),
		newConfigCommand(options),
	)
	return root
}

func newVersionCommand(info buildinfo.Info) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeVersion(cmd.OutOrStdout(), info)
		},
	}
}

func writeVersion(w io.Writer, info buildinfo.Info) error {
	_, err := fmt.Fprintf(w, "hosta %s\ncommit: %s\nbuilt: %s\n", info.Version, info.Commit, info.Date)
	return err
}
