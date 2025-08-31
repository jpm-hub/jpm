package setup

import (
	COM "jpm/common"
	DAMD64 "jpm/setup/darwin_amd64"
	DARM64 "jpm/setup/darwin_arm64"
	LAARCH64 "jpm/setup/linux_aarch64"
	LAMD64 "jpm/setup/linux_amd64"
	WAMD64 "jpm/setup/windows_amd64"
	WARM64 "jpm/setup/windows_arm64"
	"os"
	"path/filepath"
	"runtime"
)

func Setup(setupSwitch string) {
	osArch := runtime.GOOS + "-" + runtime.GOARCH
	homeDir := filepath.Join(COM.HomeDir())
	appHomeDir := COM.HomeDir()

	// Create app home directory if it doesn't exist
	os.MkdirAll(appHomeDir, 0755)

	switch osArch {
	case "linux-arm64":
		LAARCH64.SetupLinuxAarch64(setupSwitch, homeDir)
	case "linux-amd64":
		LAMD64.SetupLinuxAmd64(setupSwitch, homeDir)
	case "darwin-arm64":
		DARM64.SetupDarwinArm64(setupSwitch, homeDir)
	case "darwin-amd64":
		DAMD64.SetupDarwinAmd64(setupSwitch, homeDir)
	case "windows-arm64":
		WARM64.SetupWindowsArm64(setupSwitch, homeDir)
	case "windows-amd64":
		WAMD64.SetupWindowsAmd64(setupSwitch, homeDir)
	default:
		println("Unknown OS and ARCH")
		os.Exit(1)
	}
}
