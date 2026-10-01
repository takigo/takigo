//go:build linux

package displaylock

import (
	"os/exec"
	"syscall"
)

// dieWithParent makes the kernel kill the child when the test binary exits,
// however it exits.
func dieWithParent(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
}
