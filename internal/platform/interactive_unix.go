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

// ClearAfterConnectCommand is executed by OpenSSH's LocalCommand only after
// authentication and session establishment succeed.
func ClearAfterConnectCommand() string {
	return `printf '\033[2J\033[H'`
}
