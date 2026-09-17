package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/HobaiRiku/hosta/internal/app"
	"github.com/HobaiRiku/hosta/internal/host"
	"github.com/HobaiRiku/hosta/internal/launcher"
	"github.com/HobaiRiku/hosta/internal/openssh"
	"github.com/HobaiRiku/hosta/internal/sshconfig"
	"github.com/spf13/cobra"
)

func TestVersionCommand(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got, want := stdout.String(), "version=dev commit=unknown buildDate=unknown\n"; got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}
}

type fakeSSHClient struct {
	resolved       openssh.Resolved
	resolveAlias   string
	connectAlias   string
	explicitConfig bool
}

func (f *fakeSSHClient) Binary() string { return "/usr/bin/ssh" }
func (f *fakeSSHClient) Version(context.Context) (string, error) {
	return "OpenSSH_test", nil
}
func (f *fakeSSHClient) Resolve(_ context.Context, alias, _ string, explicit bool) (openssh.Resolved, error) {
	f.resolveAlias = alias
	f.explicitConfig = explicit
	return f.resolved, nil
}
func (f *fakeSSHClient) Connect(_ context.Context, alias, _ string, explicit bool) error {
	f.connectAlias = alias
	f.explicitConfig = explicit
	return nil
}

func TestListJSON(t *testing.T) {
	ssh := &fakeSSHClient{}
	deps := testDependencies(t, ssh)
	stdout, err := executeForTest(deps, "list", "--json")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := "[\n  {\n    \"alias\": \"home\",\n    \"displayName\": \"Home Server\",\n    \"group\": \"personal\",\n    \"hostName\": \"preview.example.com\",\n    \"user\": \"root\",\n    \"port\": \"22\",\n    \"origin\": \"native\"\n  }\n]\n"
	if stdout != want {
		t.Fatalf("list output = %q, want %q", stdout, want)
	}
}

func TestShowResolvesEffectiveConfig(t *testing.T) {
	ssh := &fakeSSHClient{resolved: openssh.Resolved{HostName: "effective.example.com", User: "root", Port: "2222"}}
	deps := testDependencies(t, ssh)
	stdout, err := executeForTest(deps, "--config", "/tmp/custom", "show", "HOME")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if ssh.resolveAlias != "HOME" || !ssh.explicitConfig {
		t.Fatalf("resolve call alias = %q, explicit = %v", ssh.resolveAlias, ssh.explicitConfig)
	}
	if !strings.Contains(stdout, "effective.example.com") || !strings.Contains(stdout, "/tmp/config:3") || strings.Contains(stdout, "Tags") {
		t.Fatalf("show output = %q", stdout)
	}
}

func TestConnectPreservesRequestedAlias(t *testing.T) {
	ssh := &fakeSSHClient{}
	deps := testDependencies(t, ssh)
	if _, err := executeForTest(deps, "connect", "HOME"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if ssh.connectAlias != "HOME" || ssh.explicitConfig {
		t.Fatalf("connect call alias = %q, explicit = %v", ssh.connectAlias, ssh.explicitConfig)
	}
}

func TestDoctorPrintsDiagnosticsAndReturnsConfigCode(t *testing.T) {
	ssh := &fakeSSHClient{}
	deps := testDependencies(t, ssh)
	deps.load = func(string) (*app.Snapshot, error) {
		index, _ := host.NewIndex(nil)
		return &app.Snapshot{
			Index: index,
			Config: &sshconfig.Config{Entry: "/tmp/config", Diagnostics: []sshconfig.Diagnostic{{
				Severity: sshconfig.SeverityError,
				Code:     "include-cycle",
				Message:  "cycle",
				Source:   sshconfig.SourceLocation{File: "/tmp/config", Line: 2},
			}}},
		}, nil
	}
	stdout, err := executeForTest(deps, "doctor")
	if err == nil || ExitCode(err) != 3 {
		t.Fatalf("doctor error = %v, code = %d", err, ExitCode(err))
	}
	if !strings.Contains(stdout, "include-cycle") || !strings.Contains(stdout, "/tmp/config:2") {
		t.Fatalf("doctor output = %q", stdout)
	}
}

func TestConfigCommand(t *testing.T) {
	stdout, err := executeForTest(testDependencies(t, &fakeSSHClient{}), "--config", "/tmp/custom", "config")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout != "/tmp/custom\n" {
		t.Fatalf("config output = %q", stdout)
	}
}

func TestCompletionScripts(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			stdout, err := executeForTest(testDependencies(t, &fakeSSHClient{}), "completion", shell, "--stdout")
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if !strings.Contains(strings.ToLower(stdout), "hosta") {
				t.Fatalf("completion output does not identify hosta: %q", stdout)
			}
		})
	}
}

func TestInstallCompletionCreatesAndUpdatesStartupFile(t *testing.T) {
	home := t.TempDir()
	startupFile := filepath.Join(home, ".zshrc")
	if err := os.WriteFile(startupFile, []byte("export EDITOR=vim\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := newRootCommand(testDependencies(t, &fakeSSHClient{}), filepath.Join(home, ".ssh", "config"))
	completionFile, updatedStartup, err := installCompletion(root, "zsh", home)
	if err != nil {
		t.Fatal(err)
	}
	if completionFile != filepath.Join(home, ".hosta_completion_zsh") || updatedStartup != startupFile {
		t.Fatalf("install target = %q, %q", completionFile, updatedStartup)
	}
	script, err := os.ReadFile(completionFile)
	if err != nil || !strings.Contains(string(script), "hosta") {
		t.Fatalf("completion script = %q, %v", script, err)
	}
	startup, err := os.ReadFile(startupFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(startup), "export EDITOR=vim") || !strings.Contains(string(startup), "source \"$HOME/.hosta_completion_zsh\"") {
		t.Fatalf("startup file = %q", startup)
	}
	if _, _, err := installCompletion(root, "zsh", home); err != nil {
		t.Fatal(err)
	}
	startup, err = os.ReadFile(startupFile)
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(string(startup), completionBlockStart); count != 1 {
		t.Fatalf("completion blocks = %d, startup = %q", count, startup)
	}
}

func TestHostCompletionUsesCatalogOnly(t *testing.T) {
	ssh := &fakeSSHClient{}
	deps := testDependencies(t, ssh)
	command := newConnectCommand(deps, &rootOptions{configPath: "/tmp/config"})
	values, directive := command.ValidArgsFunction(command, nil, "ho")
	if directive != cobra.ShellCompDirectiveNoFileComp || !reflect.DeepEqual(values, []string{"home\tHome Server"}) {
		t.Fatalf("completion = %#v, %v", values, directive)
	}
	if ssh.resolveAlias != "" || ssh.connectAlias != "" {
		t.Fatalf("completion invoked OpenSSH: %#v", ssh)
	}
}

func TestHostCompletionIncludesAlternateAliases(t *testing.T) {
	ssh := &fakeSSHClient{}
	index, err := host.NewIndex([]host.Host{{Alias: "home", Aliases: []string{"home-alt"}, DisplayName: "Home Server"}})
	if err != nil {
		t.Fatal(err)
	}
	deps := dependencies{
		load: func(string) (*app.Snapshot, error) {
			return &app.Snapshot{Index: index, Config: &sshconfig.Config{Entry: "/tmp/config"}}, nil
		},
		newSSH: func() (sshClient, error) { return ssh, nil },
	}
	command := newConnectCommand(deps, &rootOptions{configPath: "/tmp/config"})
	values, directive := command.ValidArgsFunction(command, nil, "home")
	if directive != cobra.ShellCompDirectiveNoFileComp || !reflect.DeepEqual(values, []string{"home\tHome Server", "home-alt\tHome Server"}) {
		t.Fatalf("completion = %#v, %v", values, directive)
	}
}

func TestConnectUsesRequestedAlternateAlias(t *testing.T) {
	ssh := &fakeSSHClient{}
	index, err := host.NewIndex([]host.Host{{Alias: "home", Aliases: []string{"home-alt"}}})
	if err != nil {
		t.Fatal(err)
	}
	deps := dependencies{
		load: func(string) (*app.Snapshot, error) {
			return &app.Snapshot{Index: index, Config: &sshconfig.Config{Entry: "/tmp/config"}}, nil
		},
		newSSH: func() (sshClient, error) { return ssh, nil },
	}
	if _, err := executeForTest(deps, "connect", "home-alt"); err != nil {
		t.Fatal(err)
	}
	if ssh.connectAlias != "home-alt" {
		t.Fatalf("connect alias = %q", ssh.connectAlias)
	}
}

func TestExitCode(t *testing.T) {
	if got := ExitCode(withCode(4, errors.New("missing"))); got != 4 {
		t.Fatalf("ExitCode() = %d, want 4", got)
	}
}

func TestRootRejectsNonTTY(t *testing.T) {
	_, err := executeForTest(testDependencies(t, &fakeSSHClient{}))
	if !errors.Is(err, launcher.ErrNoTTY) {
		t.Fatalf("root error = %v, want ErrNoTTY", err)
	}
}

func testDependencies(t *testing.T, ssh *fakeSSHClient) dependencies {
	t.Helper()
	index, err := host.NewIndex([]host.Host{{
		Alias:       "home",
		DisplayName: "Home Server",
		Group:       "personal",
		Preview:     host.Preview{HostName: "preview.example.com", User: "root", Port: "22"},
		Sources:     []host.Source{{File: "/tmp/config", Line: 3}},
		Origin:      host.OriginNative,
	}})
	if err != nil {
		t.Fatal(err)
	}
	return dependencies{
		load: func(string) (*app.Snapshot, error) {
			return &app.Snapshot{Index: index, Config: &sshconfig.Config{Entry: "/tmp/config"}}, nil
		},
		newSSH: func() (sshClient, error) { return ssh, nil },
	}
}

func executeForTest(deps dependencies, args ...string) (string, error) {
	var stdout bytes.Buffer
	command := newRootCommand(deps, "/tmp/config")
	command.SetOut(&stdout)
	command.SetErr(&bytes.Buffer{})
	command.SetArgs(args)
	err := command.Execute()
	return stdout.String(), err
}

func TestListTableOrder(t *testing.T) {
	stdout, err := executeForTest(testDependencies(t, &fakeSSHClient{}), "list")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 || !reflect.DeepEqual(strings.Fields(lines[0]), []string{"ALIAS", "NAME", "GROUP", "HOST"}) {
		t.Fatalf("table output = %q", stdout)
	}
}

func TestVersionFlag(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got, want := stdout.String(), "version=dev commit=unknown buildDate=unknown\n"; got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}
}

func TestShortVersionFlag(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"-v"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got, want := stdout.String(), "version=dev commit=unknown buildDate=unknown\n"; got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}
}
