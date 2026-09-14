package app

import (
	"os"
	"path/filepath"
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
