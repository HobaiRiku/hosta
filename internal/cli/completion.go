package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newCompletionCommand() *cobra.Command {
	var stdout bool
	command := &cobra.Command{
		Use:       "completion <shell>",
		Short:     "Install or print shell completion",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if stdout {
				if err := generateCompletion(cmd.Root(), cmd.OutOrStdout(), args[0]); err != nil {
					return withCode(2, err)
				}
				return nil
			}
			home, err := os.UserHomeDir()
			if err != nil {
				return withCode(3, fmt.Errorf("find home directory: %w", err))
			}
			completionFile, startupFile, err := installCompletion(cmd.Root(), args[0], home)
			if err != nil {
				return withCode(3, err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Installed %s completion at %s\nUpdated %s\n", strings.ToLower(args[0]), completionFile, startupFile)
			return err
		},
	}
	command.Flags().BoolVar(&stdout, "stdout", false, "print the completion script instead of installing it")
	return command
}

const (
	completionBlockStart = "# >>> hosta completion >>>"
	completionBlockEnd   = "# <<< hosta completion <<<"
)

type completionTarget struct {
	completionFile string
	startupFile    string
	startupBlock   string
}

func installCompletion(root *cobra.Command, shell, home string) (string, string, error) {
	target, err := completionInstallTarget(shell, home)
	if err != nil {
		return "", "", err
	}
	var script bytes.Buffer
	if err := generateCompletion(root, &script, shell); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(target.completionFile, script.Bytes(), 0o644); err != nil {
		return "", "", fmt.Errorf("write completion script %s: %w", target.completionFile, err)
	}
	if err := updateStartupFile(target.startupFile, target.startupBlock); err != nil {
		return "", "", err
	}
	return target.completionFile, target.startupFile, nil
}

func completionInstallTarget(shell, home string) (completionTarget, error) {
	shell = strings.ToLower(strings.TrimSpace(shell))
	completionFile := filepath.Join(home, ".hosta_completion_"+shell)
	switch shell {
	case "bash":
		return completionTarget{
			completionFile: completionFile,
			startupFile:    filepath.Join(home, ".bashrc"),
			startupBlock: shellSourceBlock(
				"if [ -f \"$HOME/.hosta_completion_bash\" ]; then\n  source \"$HOME/.hosta_completion_bash\"\nfi",
			),
		}, nil
	case "zsh":
		return completionTarget{
			completionFile: completionFile,
			startupFile:    filepath.Join(home, ".zshrc"),
			startupBlock: shellSourceBlock(
				"if [ -f \"$HOME/.hosta_completion_zsh\" ]; then\n  source \"$HOME/.hosta_completion_zsh\"\nfi",
			),
		}, nil
	case "fish":
		return completionTarget{
			completionFile: completionFile,
			startupFile:    filepath.Join(home, ".config", "fish", "config.fish"),
			startupBlock: shellSourceBlock(
				"if test -f ~/.hosta_completion_fish\n  source ~/.hosta_completion_fish\nend",
			),
		}, nil
	case "powershell":
		return completionTarget{
			completionFile: completionFile,
			startupFile:    filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"),
			startupBlock: shellSourceBlock(
				"if (Test-Path \"$HOME/.hosta_completion_powershell\") {\n  . \"$HOME/.hosta_completion_powershell\"\n}",
			),
		}, nil
	default:
		return completionTarget{}, fmt.Errorf("unsupported shell %q; expected bash, zsh, fish, or powershell", shell)
	}
}

func shellSourceBlock(source string) string {
	return completionBlockStart + "\n" + source + "\n" + completionBlockEnd
}

func updateStartupFile(path, block string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create completion startup directory: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read shell startup file %s: %w", path, err)
	}
	newline := "\n"
	if strings.Contains(string(data), "\r\n") {
		newline = "\r\n"
	}
	content := removeCompletionBlock(string(data))
	content = strings.TrimRight(content, "\r\n")
	if content != "" {
		content += newline + newline
	}
	content += strings.ReplaceAll(block, "\n", newline) + newline
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("update shell startup file %s: %w", path, err)
	}
	return nil
}

func removeCompletionBlock(content string) string {
	start := strings.Index(content, completionBlockStart)
	if start < 0 {
		return content
	}
	end := strings.Index(content[start:], completionBlockEnd)
	if end < 0 {
		return content
	}
	end += start + len(completionBlockEnd)
	if strings.HasPrefix(content[end:], "\r\n") {
		end += len("\r\n")
	} else if strings.HasPrefix(content[end:], "\n") {
		end++
	}
	return content[:start] + content[end:]
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
