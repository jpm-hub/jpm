//go:build !windows

package watch

import (
	"os"
	"os/exec"
	"syscall"
)

func killProcessGroup(proc *os.Process) {
	// On Unix, kill the process group
	syscall.Kill(-proc.Pid, syscall.SIGKILL)
}

func setProcessGroupAttr(cmd *exec.Cmd) {
	// Set process group ID for Unix systems
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
