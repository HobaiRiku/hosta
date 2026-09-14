package cli

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

type codedError struct {
	code int
	err  error
}

func (e codedError) Error() string { return e.err.Error() }
func (e codedError) Unwrap() error { return e.err }
func (e codedError) ExitCode() int { return e.code }

func withCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return codedError{code: code, err: err}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var coded interface{ ExitCode() int }
	if errors.As(err, &coded) {
		return coded.ExitCode()
	}
	var processError *exec.ExitError
	if errors.As(err, &processError) {
		return processError.ExitCode()
	}
	return 1
}

func exactArgs(count int) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != count {
			return withCode(2, fmt.Errorf("expected %d argument(s), received %d", count, len(args)))
		}
		return nil
	}
}
