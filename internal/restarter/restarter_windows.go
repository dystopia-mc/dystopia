//go:build windows

package restarter

import "syscall"

func init() {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
}
