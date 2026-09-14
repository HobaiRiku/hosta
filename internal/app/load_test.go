package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HobaiRiku/hosta/internal/sshconfig"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("Host home\n  HostName home.example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := Load(path, sshconfig.Options{HomeDir: dir, ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if snapshot.Index.Len() != 1 || snapshot.HasErrors() {
		t.Fatalf("snapshot = %#v", snapshot)
	}
}

func TestLoadMigratesLegacyMetadataGuard(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("Host home\n  DisplayName Home Server\n  Group personal\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := Load(path, sshconfig.Options{HomeDir: dir, ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	value, ok := snapshot.Index.Get("home")
	if !ok || value.DisplayName != "Home Server" || value.Group != "personal" {
		t.Fatalf("legacy host = %#v, %v", value, ok)
	}
	data, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(data), "IgnoreUnknown DisplayName,Group,Tags,Description\n") {
		t.Fatalf("config = %q", data)
	}
}
