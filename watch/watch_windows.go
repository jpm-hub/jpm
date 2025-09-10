//go:build windows

package watch

import (
	"os"
	"os/exec"
	"strconv"
)

func killProcessGroup(proc *os.Process) {
	// On Windows, use taskkill to kill the process tree
	exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(proc.Pid)).Run()
}

func setProcessGroupAttr(cmd *exec.Cmd) {
	// Windows doesn't need process group attributes
	// The process group is handled differently on Windows
}
