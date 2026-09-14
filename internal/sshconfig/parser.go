package sshconfig

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const defaultMaxDepth = 16

type Options struct {
	HomeDir        string
	ConfigDir      string
	LocalUser      string
	LocalUserID    string
	LocalHost      string
	MaxDepth       int
	LookupEnv      func(string) (string, bool)
	LookupUserHome func(string) (string, error)
}

type parser struct {
	options     Options
	config      *Config
	files       map[string]struct{}
	stack       map[string]struct{}
	nextScopeID int
}

func Parse(path string, options Options) (*Config, error) {
	options, err := normalizeOptions(options)
	if err != nil {
		return nil, err
	}
	entry, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}
	entry = filepath.Clean(entry)

	p := &parser{
		options: options,
		config:  &Config{Entry: entry},
		files:   make(map[string]struct{}),
		stack:   make(map[string]struct{}),
	}
	_, err = p.parseFile(entry, Scope{Kind: ScopeGlobal}, 0)
	if err != nil {
		return nil, err
	}
	p.diagnoseDuplicateAliases()
	return p.config, nil
}

func normalizeOptions(options Options) (Options, error) {
	var err error
	if options.HomeDir == "" {
		options.HomeDir, err = os.UserHomeDir()
		if err != nil {
			return options, fmt.Errorf("find home directory: %w", err)
		}
	}
	if options.ConfigDir == "" {
		options.ConfigDir = filepath.Join(options.HomeDir, ".ssh")
	}
	if options.LocalHost == "" {
		options.LocalHost, _ = os.Hostname()
	}
	if options.LocalUser == "" || options.LocalUserID == "" {
		if account, accountErr := user.Current(); accountErr == nil {
			if options.LocalUser == "" {
				options.LocalUser = account.Username
			}
			if options.LocalUserID == "" {
				options.LocalUserID = account.Uid
			}
		}
	}
	if options.MaxDepth <= 0 {
		options.MaxDepth = defaultMaxDepth
	}
	if options.LookupEnv == nil {
		options.LookupEnv = os.LookupEnv
	}
	if options.LookupUserHome == nil {
		options.LookupUserHome = lookupUserHome
	}
	return options, nil
}

func (p *parser) parseFile(path string, scope Scope, depth int) (Scope, error) {
	if depth > p.options.MaxDepth {
		p.addDiagnostic(SeverityError, "include-depth", "maximum include depth exceeded", SourceLocation{File: path})
		return scope, nil
	}

	identity := canonicalPath(path)
	if _, exists := p.stack[identity]; exists {
		p.addDiagnostic(SeverityError, "include-cycle", "recursive Include cycle detected", SourceLocation{File: path})
		return scope, nil
	}

	file, err := os.Open(path)
	if err != nil {
		if depth == 0 {
			return scope, fmt.Errorf("open SSH config %s: %w", path, err)
		}
		p.addDiagnostic(SeverityWarning, "include-unreadable", err.Error(), SourceLocation{File: path})
		return scope, nil
	}
	defer file.Close()
	if info, statErr := file.Stat(); statErr == nil && runtime.GOOS != "windows" && info.Mode().Perm()&0o022 != 0 {
		p.addDiagnostic(SeverityWarning, "insecure-permissions", "SSH config is writable by group or others", SourceLocation{File: path})
	}

	p.stack[identity] = struct{}{}
	defer delete(p.stack, identity)
	if _, exists := p.files[identity]; !exists {
		p.files[identity] = struct{}{}
		p.config.Files = append(p.config.Files, ConfigFile{Path: path})
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		source := SourceLocation{File: path, Line: lineNumber}
		directive, args, ok, parseErr := parseLine(scanner.Text())
		if parseErr != nil {
			p.addDiagnostic(SeverityError, "invalid-syntax", parseErr.Error(), source)
			continue
		}
		if !ok {
			continue
		}

		switch directive {
		case "host":
			if len(args) == 0 {
				p.addDiagnostic(SeverityError, "invalid-arguments", "Host requires at least one pattern", source)
			}
			p.nextScopeID++
			scope = Scope{ID: p.nextScopeID, Kind: ScopeHost, Patterns: append([]string(nil), args...)}
		case "match":
			if len(args) == 0 {
				p.addDiagnostic(SeverityError, "invalid-arguments", "Match requires conditions", source)
			}
			p.nextScopeID++
			scope = Scope{ID: p.nextScopeID, Kind: ScopeMatch, Conditions: append([]string(nil), args...)}
		}

		p.config.Nodes = append(p.config.Nodes, Node{
			Source:    source,
			Directive: directive,
			Args:      append([]string(nil), args...),
			Scope:     cloneScope(scope),
		})

		if directive == "include" {
			scope = p.parseIncludes(args, source, scope, depth)
		}
	}
	if err := scanner.Err(); err != nil {
		return scope, fmt.Errorf("read SSH config %s: %w", path, err)
	}
	return scope, nil
}

func (p *parser) parseIncludes(patterns []string, source SourceLocation, scope Scope, depth int) Scope {
	if len(patterns) == 0 {
		p.addDiagnostic(SeverityError, "include-empty", "Include requires at least one path", source)
		return scope
	}
	for _, rawPattern := range patterns {
		edge := IncludeEdge{Source: source, Pattern: rawPattern}
		expanded, err := p.expandInclude(rawPattern)
		if err != nil {
			p.addDiagnostic(SeverityWarning, "include-expansion", err.Error(), source)
			p.config.Includes = append(p.config.Includes, edge)
			continue
		}
		matches, err := filepath.Glob(expanded)
		if err != nil {
			p.addDiagnostic(SeverityError, "include-glob", err.Error(), source)
			p.config.Includes = append(p.config.Includes, edge)
			continue
		}
		sort.Strings(matches)
		edge.Resolved = append(edge.Resolved, matches...)
		p.config.Includes = append(p.config.Includes, edge)
		if len(matches) == 0 {
			p.addDiagnostic(SeverityWarning, "include-not-found", fmt.Sprintf("Include pattern %q matched no files", rawPattern), source)
			continue
		}
		for _, match := range matches {
			var err error
			scope, err = p.parseFile(match, scope, depth+1)
			if err != nil {
				p.addDiagnostic(SeverityWarning, "include-read", err.Error(), source)
			}
		}
	}
	return scope
}

func (p *parser) addDiagnostic(severity Severity, code, message string, source SourceLocation) {
	p.config.Diagnostics = append(p.config.Diagnostics, Diagnostic{
		Severity: severity,
		Code:     code,
		Message:  message,
		Source:   source,
	})
}

func (p *parser) diagnoseDuplicateAliases() {
	seen := make(map[string]SourceLocation)
	for _, node := range p.config.Nodes {
		if node.Directive != "host" {
			continue
		}
		for _, pattern := range node.Args {
			if !isExplicitAlias(pattern) {
				continue
			}
			key := strings.ToLower(pattern)
			if first, exists := seen[key]; exists {
				p.addDiagnostic(
					SeverityWarning,
					"duplicate-alias",
					fmt.Sprintf("Host alias %q was first declared at %s:%d", pattern, first.File, first.Line),
					node.Source,
				)
				continue
			}
			seen[key] = node.Source
		}
	}
}

func canonicalPath(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved
	}
	absolute, err := filepath.Abs(path)
	if err == nil {
		return filepath.Clean(absolute)
	}
	return filepath.Clean(path)
}

func cloneScope(scope Scope) Scope {
	scope.Patterns = append([]string(nil), scope.Patterns...)
	scope.Conditions = append([]string(nil), scope.Conditions...)
	return scope
}
