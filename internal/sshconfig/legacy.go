package sshconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const legacyIgnoreUnknown = "IgnoreUnknown DisplayName,Group,Tags,Description"

var legacyMetadataDirectives = map[string]struct{}{
	"displayname": {},
	"group":       {},
	"tags":        {},
	"description": {},
}

// EnsureLegacyCompatibility protects old, directive-based Hosta metadata from
// OpenSSH parse errors. New # @hosta.* annotations never require this guard.
func EnsureLegacyCompatibility(config *Config) (bool, error) {
	if config == nil || !needsLegacyIgnore(config.Nodes) {
		return false, nil
	}

	data, err := os.ReadFile(config.Entry)
	if err != nil {
		return false, fmt.Errorf("read SSH config for legacy compatibility: %w", err)
	}
	newline := []byte("\n")
	if strings.Contains(string(data), "\r\n") {
		newline = []byte("\r\n")
	}
	updated := make([]byte, 0, len(data)+len(legacyIgnoreUnknown)+len(newline))
	updated = append(updated, legacyIgnoreUnknown...)
	updated = append(updated, newline...)
	updated = append(updated, data...)

	if err := replaceFile(config.Entry, updated); err != nil {
		return false, fmt.Errorf("add IgnoreUnknown guard to SSH config: %w", err)
	}
	return true, nil
}

func needsLegacyIgnore(nodes []Node) bool {
	ignored := make(map[string]bool)
	for _, node := range nodes {
		if node.Directive == "ignoreunknown" {
			for _, arg := range node.Args {
				for _, pattern := range strings.Split(arg, ",") {
					pattern = strings.ToLower(strings.TrimSpace(pattern))
					for directive := range legacyMetadataDirectives {
						if matched, _ := filepath.Match(pattern, directive); matched {
							ignored[directive] = true
						}
					}
				}
			}
			continue
		}
		if _, legacy := legacyMetadataDirectives[node.Directive]; legacy && !ignored[node.Directive] {
			return true
		}
	}
	return false
}

func replaceFile(path string, data []byte) error {
	target := path
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		target = resolved
	}
	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(target), ".hosta-config-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return replacePath(temporaryPath, target)
}
