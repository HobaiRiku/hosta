package sshconfig

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseIncludesAndDiscoverHosts(t *testing.T) {
	sshDir := t.TempDir()
	mustWrite(t, filepath.Join(sshDir, "config"), `
Include conf.d/*.conf

Host *.example.com !blocked.example.com
    User deploy

Host home
    # @hosta.display-name Home Server
    # @hosta.group personal
    # @hosta.tags home server
    HostName home.example.com
`)
	mustWrite(t, filepath.Join(sshDir, "conf.d", "20-work.conf"), `
Host dev prod
    # @hosta.display-name Work
    # @hosta.group work
    # @hosta.tags linux
Include shared.conf
`)
	mustWrite(t, filepath.Join(sshDir, "conf.d", "10-first.conf"), "Host alpha\n")
	mustWrite(t, filepath.Join(sshDir, "shared.conf"), `
Host nas
    # @hosta.description "Storage #1"
`)

	config, err := Parse(filepath.Join(sshDir, "config"), Options{
		HomeDir:   filepath.Dir(sshDir),
		ConfigDir: sshDir,
		LocalUser: "tester",
		LocalHost: "workstation",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(config.Files) != 4 {
		t.Fatalf("files = %#v, want 4", config.Files)
	}
	if got := filepath.Base(config.Files[1].Path); got != "10-first.conf" {
		t.Fatalf("first glob match = %q, want 10-first.conf", got)
	}

	hosts := config.Hosts()
	aliases := make([]string, len(hosts))
	for i := range hosts {
		aliases[i] = hosts[i].Alias
	}
	if want := []string{"alpha", "dev", "home", "nas", "prod"}; !reflect.DeepEqual(aliases, want) {
		t.Fatalf("aliases = %#v, want %#v", aliases, want)
	}
	dev := hosts[1]
	if dev.DisplayName != "Work" || dev.Group != "work" || !reflect.DeepEqual(dev.Tags, []string{"linux"}) {
		t.Fatalf("dev metadata = %#v", dev)
	}
	home := hosts[2]
	if home.Sources[0].Line != 7 || home.DisplayName != "Home Server" || home.HostName != "home.example.com" {
		t.Fatalf("home = %#v", home)
	}
	if hosts[3].Description != "Storage #1" {
		t.Fatalf("nas = %#v", hosts[3])
	}
}

func TestParseReportsIncludeCycleAndMissingFile(t *testing.T) {
	sshDir := t.TempDir()
	mustWrite(t, filepath.Join(sshDir, "config"), "Include a.conf missing.conf\n")
	mustWrite(t, filepath.Join(sshDir, "a.conf"), "Include b.conf\n")
	mustWrite(t, filepath.Join(sshDir, "b.conf"), "Include a.conf\n")

	config, err := Parse(filepath.Join(sshDir, "config"), Options{HomeDir: t.TempDir(), ConfigDir: sshDir})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	codes := make(map[string]bool)
	for _, diagnostic := range config.Diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, want := range []string{"include-cycle", "include-not-found"} {
		if !codes[want] {
			t.Fatalf("diagnostics = %#v, want %q", config.Diagnostics, want)
		}
	}
}

func TestParseExpandsEnvironmentHomeAndTokens(t *testing.T) {
	root := t.TempDir()
	sshDir := filepath.Join(root, ".ssh")
	shared := filepath.Join(root, "shared")
	mustWrite(t, filepath.Join(sshDir, "config"), "Include ${SHARED}/*.conf ~/personal.conf %d/token.conf %d/%i.conf\n")
	mustWrite(t, filepath.Join(shared, "a.conf"), "Host shared\n")
	mustWrite(t, filepath.Join(root, "personal.conf"), "Host personal\n")
	mustWrite(t, filepath.Join(root, "token.conf"), "Host token\n")
	mustWrite(t, filepath.Join(root, "501.conf"), "Host uid-token\n")

	config, err := Parse(filepath.Join(sshDir, "config"), Options{
		HomeDir:     root,
		ConfigDir:   sshDir,
		LocalUserID: "501",
		LookupEnv: func(name string) (string, bool) {
			if name == "SHARED" {
				return shared, true
			}
			return "", false
		},
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	aliases := []string{}
	for _, host := range config.Hosts() {
		aliases = append(aliases, host.Alias)
	}
	if want := []string{"personal", "shared", "token", "uid-token"}; !reflect.DeepEqual(aliases, want) {
		t.Fatalf("aliases = %#v, want %#v", aliases, want)
	}
}

func TestParseReportsMissingHostArguments(t *testing.T) {
	sshDir := t.TempDir()
	mustWrite(t, filepath.Join(sshDir, "config"), "Host\n")

	config, err := Parse(filepath.Join(sshDir, "config"), Options{HomeDir: t.TempDir(), ConfigDir: sshDir})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(config.Diagnostics) != 1 || config.Diagnostics[0].Code != "invalid-arguments" {
		t.Fatalf("diagnostics = %#v, want invalid-arguments", config.Diagnostics)
	}
}

func TestParseReportsDuplicateAlias(t *testing.T) {
	sshDir := t.TempDir()
	mustWrite(t, filepath.Join(sshDir, "config"), "Host home\nHost HOME\n")

	config, err := Parse(filepath.Join(sshDir, "config"), Options{HomeDir: t.TempDir(), ConfigDir: sshDir})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !hasDiagnostic(config.Diagnostics, "duplicate-alias") {
		t.Fatalf("diagnostics = %#v, want duplicate-alias", config.Diagnostics)
	}
	hosts := config.Hosts()
	if len(hosts) != 1 || len(hosts[0].Sources) != 2 {
		t.Fatalf("hosts = %#v, want one host with two sources", hosts)
	}
}

func hasDiagnostic(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func TestParseReportsDynamicIncludeToken(t *testing.T) {
	sshDir := t.TempDir()
	mustWrite(t, filepath.Join(sshDir, "config"), "Include hosts/%h.conf\nHost home\n")

	config, err := Parse(filepath.Join(sshDir, "config"), Options{HomeDir: t.TempDir(), ConfigDir: sshDir})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(config.Diagnostics) != 1 || config.Diagnostics[0].Code != "include-expansion" {
		t.Fatalf("diagnostics = %#v, want include-expansion", config.Diagnostics)
	}
}

func TestParseRootMissingIsFatal(t *testing.T) {
	_, err := Parse(filepath.Join(t.TempDir(), "missing"), Options{})
	if err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
