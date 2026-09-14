package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "completion <shell>",
		Short:     "Generate shell completion script",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := generateCompletion(cmd.Root(), cmd.OutOrStdout(), args[0]); err != nil {
				return withCode(2, err)
			}
			return nil
		},
	}
}

func generateCompletion(root *cobra.Command, output io.Writer, shell string) error {
	switch strings.ToLower(shell) {
	case "bash":
		return root.GenBashCompletionV2(output, true)
	case "zsh":
		return root.GenZshCompletion(output)
	case "fish":
		return root.GenFishCompletion(output, true)
	case "powershell":
		return root.GenPowerShellCompletionWithDesc(output)
	default:
		return fmt.Errorf("unsupported shell %q; expected bash, zsh, fish, or powershell", shell)
	}
}

func completeHosts(deps dependencies, options *rootOptions) cobra.CompletionFunc {
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		snapshot, err := deps.load(options.configPath)
		if err != nil || snapshot.HasErrors() {
			return nil, cobra.ShellCompDirectiveError
		}
		hosts := snapshot.Index.All()
		values := make([]string, 0, len(hosts))
		for _, value := range hosts {
			description := strings.ReplaceAll(value.DisplayName, "\t", " ")
			description = strings.ReplaceAll(description, "\n", " ")
			for _, alias := range append([]string{value.Alias}, value.Aliases...) {
				if description == "" {
					values = append(values, alias)
				} else {
					values = append(values, alias+"\t"+description)
				}
			}
		}
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}
