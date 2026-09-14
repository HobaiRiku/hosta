package sshconfig

import (
	"reflect"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		directive string
		args      []string
		ok        bool
	}{
		{name: "blank", line: "  # comment", ok: false},
		{name: "spaces", line: "Host home dev", directive: "host", args: []string{"home", "dev"}, ok: true},
		{name: "equals", line: "Port = 2222", directive: "port", args: []string{"2222"}, ok: true},
		{name: "attached equals", line: "User=root", directive: "user", args: []string{"root"}, ok: true},
		{name: "quoted", line: `Description "Home server #1" # note`, directive: "description", args: []string{"Home server #1"}, ok: true},
		{name: "value equals", line: "SetEnv FOO=bar", directive: "setenv", args: []string{"FOO=bar"}, ok: true},
		{name: "escaped space", line: `Include cloud\ folder/*.conf`, directive: "include", args: []string{"cloud folder/*.conf"}, ok: true},
		{name: "hosta annotation", line: `  # @hosta.display-name Home Server`, directive: "hosta.display-name", args: []string{"Home", "Server"}, ok: true},
		{name: "ordinary comment", line: `# deployment hosts`, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directive, args, ok, err := parseLine(tt.line)
			if err != nil {
				t.Fatalf("parseLine() error = %v", err)
			}
			if directive != tt.directive || ok != tt.ok || !reflect.DeepEqual(args, tt.args) {
				t.Fatalf("parseLine() = %q, %#v, %v; want %q, %#v, %v", directive, args, ok, tt.directive, tt.args, tt.ok)
			}
		})
	}
}

func TestParseLineRejectsUnterminatedQuote(t *testing.T) {
	if _, _, _, err := parseLine(`Host "home`); err == nil {
		t.Fatal("parseLine() error = nil, want syntax error")
	}
}
