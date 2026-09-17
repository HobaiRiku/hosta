package sshconfig

type ScopeKind string

const (
	ScopeGlobal ScopeKind = "global"
	ScopeHost   ScopeKind = "host"
	ScopeMatch  ScopeKind = "match"
)

type SourceLocation struct {
	File string
	Line int
}

type Scope struct {
	ID         int
	Kind       ScopeKind
	Patterns   []string
	Conditions []string
}

type Node struct {
	Source    SourceLocation
	Directive string
	Args      []string
	Scope     Scope
}

type ConfigFile struct {
	Path string
}

type IncludeEdge struct {
	Source   SourceLocation
	Pattern  string
	Resolved []string
}

type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Diagnostic struct {
	Severity Severity
	Code     string
	Message  string
	Source   SourceLocation
}

type Config struct {
	Entry       string
	Files       []ConfigFile
	Nodes       []Node
	Includes    []IncludeEdge
	Diagnostics []Diagnostic
}

type Host struct {
	Alias       string
	Aliases     []string
	DisplayName string
	Group       string
	Description string
	HostName    string
	User        string
	Port        string
	Sources     []SourceLocation
}
