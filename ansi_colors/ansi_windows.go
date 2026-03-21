//go:build windows

package ansi_colors

import (
	"os"

	"golang.org/x/sys/windows"
)

func EnableANSIWindows() {
	for _, handle := range []windows.Handle{
		windows.Handle(os.Stdout.Fd()),
		windows.Handle(os.Stderr.Fd()),
	} {
		var mode uint32
		if err := windows.GetConsoleMode(handle, &mode); err != nil {
			continue
		}
		windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}
