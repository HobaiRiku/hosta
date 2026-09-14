//go:build !windows

package sshconfig

import "os"

func replacePath(source, target string) error {
	return os.Rename(source, target)
}
