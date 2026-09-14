package cli

import (
	"fmt"
	"strings"

	"github.com/HobaiRiku/hosta/internal/sshconfig"
	"github.com/spf13/cobra"
)

func newDoctorCommand(deps dependencies, options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the local OpenSSH environment and configuration",
		Args:  exactArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			writer := cmd.OutOrStdout()
			fmt.Fprint(writer, "Hosta Doctor\n\n")

			client, sshErr := deps.newSSH()
			if sshErr != nil {
				fmt.Fprintf(writer, "✗ OpenSSH\n  %v\n", sshErr)
			} else {
				version, versionErr := client.Version(cmd.Context())
				if versionErr != nil {
					fmt.Fprintf(writer, "⚠ OpenSSH\n  %s\n  %v\n", client.Binary(), versionErr)
				} else {
					fmt.Fprintf(writer, "✓ OpenSSH\n  %s\n  %s\n", client.Binary(), version)
				}
			}

			snapshot, configErr := deps.load(options.configPath)
			if configErr != nil {
				fmt.Fprintf(writer, "\n✗ SSH config\n  %s\n  %v\n", options.configPath, configErr)
			} else {
				symbol := "✓"
				if snapshot.HasErrors() {
					symbol = "✗"
				} else if len(snapshot.Config.Diagnostics) > 0 {
					symbol = "⚠"
				}
				fmt.Fprintf(writer, "\n%s SSH config\n  %s\n  %d hosts discovered\n", symbol, snapshot.Config.Entry, snapshot.Index.Len())
				for _, diagnostic := range snapshot.Config.Diagnostics {
					symbol := "⚠"
					if diagnostic.Severity == sshconfig.SeverityError {
						symbol = "✗"
					}
					location := diagnostic.Source.File
					if diagnostic.Source.Line > 0 {
						location = fmt.Sprintf("%s:%d", location, diagnostic.Source.Line)
					}
					fmt.Fprintf(writer, "%s %s\n  %s\n  %s\n", symbol, diagnostic.Code, location, strings.TrimSpace(diagnostic.Message))
				}
			}

			if sshErr != nil {
				return withCode(4, fmt.Errorf("doctor found no OpenSSH client"))
			}
			if configErr != nil || snapshot.HasErrors() {
				return withCode(3, fmt.Errorf("doctor found SSH configuration errors"))
			}
			return nil
		},
	}
}
