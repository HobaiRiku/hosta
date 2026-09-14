package sshconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureLegacyCompatibility(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	original := "Host home\n  DisplayName Home Server\n  Group personal\n  Tags home server\n  Description Primary host\n"
	mustWrite(t, path, original)

	config, err := Parse(path, Options{HomeDir: dir, ConfigDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	changed, err := EnsureLegacyCompatibility(config)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("legacy config was not guarded")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != legacyIgnoreUnknown+"\n"+original {
		t.Fatalf("config = %q", got)
	}

	config, err = Parse(path, Options{HomeDir: dir, ConfigDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	changed, err = EnsureLegacyCompatibility(config)
	if err != nil || changed {
		t.Fatalf("second guard = %v, %v", changed, err)
	}
}

func TestEnsureLegacyCompatibilityLeavesAnnotationsUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	original := "Host home\n  # @hosta.display-name Home Server\n"
	mustWrite(t, path, original)

	config, err := Parse(path, Options{HomeDir: dir, ConfigDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	changed, err := EnsureLegacyCompatibility(config)
	if err != nil || changed {
		t.Fatalf("annotation guard = %v, %v", changed, err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != original {
		t.Fatalf("config changed to %q", data)
	}
}

func TestEnsureLegacyCompatibilityPreservesCRLF(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	mustWrite(t, path, "Host home\r\n  DisplayName Home\r\n")

	config, err := Parse(path, Options{HomeDir: dir, ConfigDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureLegacyCompatibility(config); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(data), legacyIgnoreUnknown+"\r\n") {
		t.Fatalf("config does not preserve CRLF: %q", data)
	}
}
