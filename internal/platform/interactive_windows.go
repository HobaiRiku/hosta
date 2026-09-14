//go:build windows

package platform

import (
	"context"
	"os"
	"os/exec"
)

func RunInteractive(ctx context.Context, binary string, args []string) error {
	command := exec.CommandContext(ctx, binary, args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
