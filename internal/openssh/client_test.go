package openssh

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeRunner struct {
	stdout          []byte
	stderr          []byte
	err             error
	outputArgs      []string
	interactiveArgs []string
}

func (f *fakeRunner) Output(_ context.Context, _ string, args []string) ([]byte, []byte, error) {
	f.outputArgs = append([]string(nil), args...)
	return f.stdout, f.stderr, f.err
}

func (f *fakeRunner) Interactive(_ context.Context, _ string, args []string) error {
	f.interactiveArgs = append([]string(nil), args...)
	return f.err
}

func TestResolveUsesOpenSSHAndParsesFields(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("host home\nhostname home.example.com\nuser root\nport 2222\nidentityfile ~/.ssh/a\nidentityfile ~/.ssh/b\n")}
	client, err := New("/usr/bin/ssh", runner)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := client.Resolve(context.Background(), "home", "/tmp/config", true)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if resolved.HostName != "home.example.com" || resolved.User != "root" || resolved.Port != "2222" || len(resolved.IdentityFiles) != 2 {
		t.Fatalf("resolved = %#v", resolved)
	}
	wantArgs := []string{"-G", "-T", "-F", "/tmp/config", "--", "home"}
	if !reflect.DeepEqual(runner.outputArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", runner.outputArgs, wantArgs)
	}
}

func TestConnectDoesNotForceDefaultConfig(t *testing.T) {
	runner := &fakeRunner{}
	client, _ := New("ssh", runner)
	if err := client.Connect(context.Background(), "home", "/home/me/.ssh/config", false); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if want := []string{"--", "home"}; !reflect.DeepEqual(runner.interactiveArgs, want) {
		t.Fatalf("args = %#v, want %#v", runner.interactiveArgs, want)
	}
}

func TestValidateAlias(t *testing.T) {
	for _, alias := range []string{"", "-oProxyCommand=x", "two hosts", "bad\x00host"} {
		if err := ValidateAlias(alias); err == nil {
			t.Fatalf("ValidateAlias(%q) error = nil", alias)
		}
	}
	if err := ValidateAlias("开发机"); err != nil {
		t.Fatalf("ValidateAlias() error = %v", err)
	}
}

func TestResolveIncludesStderrOnFailure(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("bad configuration"), err: errors.New("exit status 255")}
	client, _ := New("ssh", runner)
	if _, err := client.Resolve(context.Background(), "home", "", false); err == nil {
		t.Fatal("Resolve() error = nil")
	}
}

func TestResolvedConnectionCommand(t *testing.T) {
	command, err := (Resolved{HostName: "192.0.2.10", User: "root", Port: "2222"}).ConnectionCommand()
	if err != nil {
		t.Fatal(err)
	}
	if want := "ssh -p 2222 root@192.0.2.10"; command != want {
		t.Fatalf("command = %q, want %q", command, want)
	}
}

func TestResolvedConnectionCommandQuotesUnsafeValues(t *testing.T) {
	command, err := (Resolved{HostName: "host name", User: "root", Port: "22"}).ConnectionCommand()
	if err != nil {
		t.Fatal(err)
	}
	if want := "ssh -p 22 'root@host name'"; command != want {
		t.Fatalf("command = %q, want %q", command, want)
	}
}
