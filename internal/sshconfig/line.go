package sshconfig

import (
	"fmt"
	"strings"
	"unicode"
)

func parseLine(line string) (string, []string, bool, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", nil, false, nil
	}

	end := 0
	for end < len(line) && line[end] != '=' && !unicode.IsSpace(rune(line[end])) {
		end++
	}
	if end == 0 {
		return "", nil, false, fmt.Errorf("missing directive")
	}

	directive := strings.ToLower(line[:end])
	rest := strings.TrimLeftFunc(line[end:], unicode.IsSpace)
	if strings.HasPrefix(rest, "=") {
		rest = strings.TrimLeftFunc(rest[1:], unicode.IsSpace)
	}

	args, err := splitArguments(rest)
	if err != nil {
		return "", nil, false, err
	}
	return directive, args, true, nil
}

func splitArguments(input string) ([]string, error) {
	var (
		args    []string
		current strings.Builder
		quote   rune
		escaped bool
		started bool
	)

	flush := func() {
		if started {
			args = append(args, current.String())
			current.Reset()
			started = false
		}
	}

	for _, char := range input {
		if escaped {
			current.WriteRune(char)
			escaped = false
			started = true
			continue
		}
		if char == '\\' && quote != '\'' {
			escaped = true
			started = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
			} else {
				current.WriteRune(char)
			}
			started = true
			continue
		}
		switch {
		case char == '\'' || char == '"':
			quote = char
			started = true
		case unicode.IsSpace(char):
			flush()
		case char == '#' && !started:
			return args, nil
		default:
			current.WriteRune(char)
			started = true
		}
	}

	if escaped {
		return nil, fmt.Errorf("trailing escape")
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quote")
	}
	flush()
	return args, nil
}
