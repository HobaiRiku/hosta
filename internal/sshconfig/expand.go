package sshconfig

import (
	"fmt"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var environmentPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func (p *parser) expandInclude(pattern string) (string, error) {
	var expansionErr error
	pattern = environmentPattern.ReplaceAllStringFunc(pattern, func(match string) string {
		name := environmentPattern.FindStringSubmatch(match)[1]
		value, ok := p.options.LookupEnv(name)
		if !ok {
			expansionErr = fmt.Errorf("environment variable %s is not set", name)
			return match
		}
		return value
	})
	if expansionErr != nil {
		return "", expansionErr
	}

	var expanded strings.Builder
	for i := 0; i < len(pattern); i++ {
		if pattern[i] != '%' {
			expanded.WriteByte(pattern[i])
			continue
		}
		if i+1 >= len(pattern) {
			return "", fmt.Errorf("incomplete token %%")
		}
		i++
		switch pattern[i] {
		case '%':
			expanded.WriteByte('%')
		case 'd':
			expanded.WriteString(p.options.HomeDir)
		case 'u':
			expanded.WriteString(p.options.LocalUser)
		case 'i':
			expanded.WriteString(p.options.LocalUserID)
		case 'l', 'L':
			expanded.WriteString(p.options.LocalHost)
		default:
			return "", fmt.Errorf("token %%%c requires a target host and cannot be discovered statically", pattern[i])
		}
	}
	pattern = expanded.String()

	if strings.HasPrefix(pattern, "~") {
		home, rest, err := p.expandHome(pattern)
		if err != nil {
			return "", err
		}
		pattern = filepath.Join(home, rest)
	}
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(p.options.ConfigDir, pattern)
	}
	return filepath.Clean(pattern), nil
}

func (p *parser) expandHome(pattern string) (string, string, error) {
	separator := strings.IndexAny(pattern, `/\\`)
	if separator == -1 {
		separator = len(pattern)
	}
	username := pattern[1:separator]
	rest := strings.TrimLeft(pattern[separator:], `/\\`)
	if username == "" {
		return p.options.HomeDir, rest, nil
	}
	home, err := p.options.LookupUserHome(username)
	if err != nil {
		return "", "", fmt.Errorf("resolve home for %q: %w", username, err)
	}
	return home, rest, nil
}

func lookupUserHome(username string) (string, error) {
	account, err := user.Lookup(username)
	if err != nil {
		return "", err
	}
	return account.HomeDir, nil
}
