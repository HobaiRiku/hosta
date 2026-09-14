package app

import (
	"testing"

	"github.com/HobaiRiku/hosta/internal/host"
	"github.com/HobaiRiku/hosta/internal/sshconfig"
)

func TestIndexFromSSHConfig(t *testing.T) {
	config := &sshconfig.Config{Nodes: []sshconfig.Node{
		{
			Source:    sshconfig.SourceLocation{File: "/tmp/config", Line: 3},
			Directive: "host",
			Args:      []string{"home"},
			Scope:     sshconfig.Scope{ID: 1, Kind: sshconfig.ScopeHost, Patterns: []string{"home"}},
		},
		{
			Source:    sshconfig.SourceLocation{File: "/tmp/config", Line: 4},
			Directive: "hostname",
			Args:      []string{"home.example.com"},
			Scope:     sshconfig.Scope{ID: 1, Kind: sshconfig.ScopeHost, Patterns: []string{"home"}},
		},
	}}

	index, err := IndexFromSSHConfig(config)
	if err != nil {
		t.Fatalf("IndexFromSSHConfig() error = %v", err)
	}
	got := index.All()
	if len(got) != 1 || got[0].ID != "native:home" || got[0].Origin != host.OriginNative || got[0].Preview.HostName != "home.example.com" {
		t.Fatalf("index hosts = %#v", got)
	}
}

func TestIndexFromSSHConfigRejectsNil(t *testing.T) {
	if _, err := IndexFromSSHConfig(nil); err == nil {
		t.Fatal("IndexFromSSHConfig(nil) error = nil, want error")
	}
}
