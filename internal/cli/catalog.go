package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/HobaiRiku/hosta/internal/app"
	"github.com/HobaiRiku/hosta/internal/host"
	"github.com/HobaiRiku/hosta/internal/openssh"
	"github.com/spf13/cobra"
)

func newListCommand(deps dependencies, options *rootOptions) *cobra.Command {
	var asJSON bool
	command := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List discovered SSH hosts",
		Args:    exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			snapshot, err := loadSnapshot(deps, options.configPath)
			if err != nil {
				return err
			}
			if asJSON {
				return writeHostJSON(cmd.OutOrStdout(), snapshot.Index.All())
			}
			return writeHostTable(cmd.OutOrStdout(), snapshot.Index.All())
		},
	}
	command.Flags().BoolVar(&asJSON, "json", false, "write stable JSON output")
	return command
}

func newShowCommand(deps dependencies, options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "show <host>",
		Short: "Show Hosta metadata and effective OpenSSH configuration",
		Args:  exactArgs(1),
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
			resolved, err := client.Resolve(cmd.Context(), value.Alias, options.configPath, configWasExplicit(cmd))
			if err != nil {
				return withCode(5, err)
			}
			return writeHostDetails(cmd.OutOrStdout(), value, resolved)
		},
	}
}

func loadSnapshot(deps dependencies, path string) (*app.Snapshot, error) {
	snapshot, err := deps.load(path)
	if err != nil {
		return nil, withCode(3, err)
	}
	if snapshot.HasErrors() {
		return nil, withCode(3, fmt.Errorf("SSH config contains errors; run `hosta doctor` for details"))
	}
	return snapshot, nil
}

type jsonHost struct {
	Alias       string   `json:"alias"`
	DisplayName string   `json:"displayName,omitempty"`
	Group       string   `json:"group,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
	HostName    string   `json:"hostName,omitempty"`
	User        string   `json:"user,omitempty"`
	Port        string   `json:"port,omitempty"`
	Origin      string   `json:"origin"`
}

func writeHostJSON(w io.Writer, hosts []host.Host) error {
	values := make([]jsonHost, len(hosts))
	for i, value := range hosts {
		values[i] = jsonHost{
			Alias:       value.Alias,
			DisplayName: value.DisplayName,
			Group:       value.Group,
			Tags:        append([]string(nil), value.Tags...),
			Description: value.Description,
			HostName:    value.Preview.HostName,
			User:        value.Preview.User,
			Port:        value.Preview.Port,
			Origin:      string(value.Origin),
		}
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(values)
}

func writeHostTable(w io.Writer, hosts []host.Host) error {
	table := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(table, "ALIAS\tNAME\tGROUP\tHOST"); err != nil {
		return err
	}
	for _, value := range hosts {
		if _, err := fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", value.Alias, value.DisplayName, value.Group, value.Preview.HostName); err != nil {
			return err
		}
	}
	return table.Flush()
}

func writeHostDetails(w io.Writer, value host.Host, resolved openssh.Resolved) error {
	name := value.DisplayName
	if name == "" {
		name = value.Alias
	}
	table := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintf(table, "%s\n\nAlias\t%s\nGroup\t%s\nTags\t%s\nDescription\t%s\n\nSSH (resolved by OpenSSH)\nHostName\t%s\nUser\t%s\nPort\t%s\n", name, value.Alias, value.Group, strings.Join(value.Tags, " "), value.Description, resolved.HostName, resolved.User, resolved.Port); err != nil {
		return err
	}
	for _, identity := range resolved.IdentityFiles {
		if _, err := fmt.Fprintf(table, "Identity\t%s\n", identity); err != nil {
			return err
		}
	}
	if err := table.Flush(); err != nil {
		return err
	}
	if len(value.Sources) > 0 {
		if _, err := fmt.Fprintln(w, "\nSource"); err != nil {
			return err
		}
		for _, source := range value.Sources {
			if _, err := fmt.Fprintf(w, "%s:%d\n", source.File, source.Line); err != nil {
				return err
			}
		}
	}
	return nil
}
