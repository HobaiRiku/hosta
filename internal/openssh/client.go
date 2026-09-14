package openssh

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"unicode"

	"github.com/HobaiRiku/hosta/internal/platform"
)

type Runner interface {
	Output(context.Context, string, []string) ([]byte, []byte, error)
	Interactive(context.Context, string, []string) error
}

type Client struct {
	binary string
	runner Runner
}

type Resolved struct {
	Host          string
	HostName      string
	User          string
	Port          string
	IdentityFiles []string
}

// ConnectionCommand returns a portable shell command using the effective
// destination values returned by ssh -G. It intentionally omits IdentityFile
// because local key paths are not safe to copy into a shareable command.
func (r Resolved) ConnectionCommand() (string, error) {
	if r.HostName == "" {
		return "", fmt.Errorf("resolved SSH host has no hostname")
	}
	target := r.HostName
	if r.User != "" {
		target = r.User + "@" + target
	}
	args := []string{"ssh"}
	if r.Port != "" {
		args = append(args, "-p", shellQuote(r.Port))
	}
	args = append(args, shellQuote(target))
	return strings.Join(args, " "), nil
}

func shellQuote(value string) string {
	if value != "" {
		safe := true
		for _, char := range value {
			if !(unicode.IsLetter(char) || unicode.IsDigit(char) || strings.ContainsRune("@%+=:,./_-", char)) {
				safe = false
				break
			}
		}
		if safe {
			return value
		}
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func New(binary string, runner Runner) (*Client, error) {
	if binary == "" {
		var err error
		binary, err = exec.LookPath("ssh")
		if err != nil {
			return nil, fmt.Errorf("find OpenSSH client: %w", err)
		}
	}
	if runner == nil {
		runner = osRunner{}
	}
	return &Client{binary: binary, runner: runner}, nil
}

func (c *Client) Binary() string {
	return c.binary
}

func (c *Client) Version(ctx context.Context) (string, error) {
	stdout, stderr, err := c.runner.Output(ctx, c.binary, []string{"-V"})
	if err != nil {
		return "", commandError("read OpenSSH version", err, stderr)
	}
	version := strings.TrimSpace(string(stderr))
	if version == "" {
		version = strings.TrimSpace(string(stdout))
	}
	return version, nil
}

func (c *Client) Resolve(ctx context.Context, alias, configPath string, explicitConfig bool) (Resolved, error) {
	if err := ValidateAlias(alias); err != nil {
		return Resolved{}, err
	}
	args := []string{"-G", "-T"}
	args = appendConfig(args, configPath, explicitConfig)
	args = append(args, "--", alias)
	stdout, stderr, err := c.runner.Output(ctx, c.binary, args)
	if err != nil {
		return Resolved{}, commandError("resolve SSH host", err, stderr)
	}
	return parseResolved(stdout), nil
}

func (c *Client) Connect(ctx context.Context, alias, configPath string, explicitConfig bool) error {
	if err := ValidateAlias(alias); err != nil {
		return err
	}
	args := appendConfig(nil, configPath, explicitConfig)
	args = append(args, "--", alias)
	if err := c.runner.Interactive(ctx, c.binary, args); err != nil {
		return fmt.Errorf("run OpenSSH: %w", err)
	}
	return nil
}

func ValidateAlias(alias string) error {
	if alias == "" {
		return fmt.Errorf("SSH host alias is empty")
	}
	if strings.HasPrefix(alias, "-") {
		return fmt.Errorf("SSH host alias %q starts with an option prefix", alias)
	}
	for _, char := range alias {
		if unicode.IsSpace(char) || unicode.IsControl(char) || char == 0 {
			return fmt.Errorf("SSH host alias %q contains whitespace or control characters", alias)
		}
	}
	return nil
}

func appendConfig(args []string, configPath string, explicit bool) []string {
	if explicit {
		return append(args, "-F", configPath)
	}
	return args
}

func parseResolved(output []byte) Resolved {
	var resolved Resolved
	for _, line := range strings.Split(string(output), "\n") {
		key, value, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.ToLower(key) {
		case "host":
			resolved.Host = value
		case "hostname":
			resolved.HostName = value
		case "user":
			resolved.User = value
		case "port":
			resolved.Port = value
		case "identityfile":
			resolved.IdentityFiles = append(resolved.IdentityFiles, value)
		}
	}
	return resolved
}

func commandError(action string, err error, stderr []byte) error {
	detail := strings.TrimSpace(string(stderr))
	if detail == "" {
		return fmt.Errorf("%s: %w", action, err)
	}
	return fmt.Errorf("%s: %w: %s", action, err, detail)
}

type osRunner struct{}

func (osRunner) Output(ctx context.Context, binary string, args []string) ([]byte, []byte, error) {
	command := exec.CommandContext(ctx, binary, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func (osRunner) Interactive(ctx context.Context, binary string, args []string) error {
	return platform.RunInteractive(ctx, binary, args)
}
