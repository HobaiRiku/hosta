//go:build !windows

package platform

import (
	"context"
	"os"
	"syscall"
)

func RunInteractive(_ context.Context, binary string, args []string) error {
	argv := append([]string{binary}, args...)
	return syscall.Exec(binary, argv, os.Environ())
}
