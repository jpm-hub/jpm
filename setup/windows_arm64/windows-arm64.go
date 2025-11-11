package windows_arm64

import (
	COM "jpm/common"
	"jpm/setup/osless"
	"os"
	"path/filepath"
	"strings"
)

func SetupWindowsArm64(setupSwitch string, homeDir string) {
	switch setupSwitch {
	case "-hotswap":
		if COM.JAVA() == "java" {
			println("\n Please setup dcevm first with : 'jpm setup -java'\n")
			return
		}
		osless.HotSwapAgentPlugin(homeDir)
	case "-java":
		dcevmPlugin(homeDir)
	case "-junit":
		osless.JunitPlugin(homeDir)
	case "-kotlin":
		osless.Kotlinc(homeDir)
	}
}

func dcevmPlugin(homeDir string) {
	// DCEVM for Windows ARM64
	url := "https://cache-redirector.jetbrains.com/intellij-jbr/jbr-25-windows-aarch64-b176.4.tar.gz"
	filename := "jbr-25-windows-aarch64-b176.4.tar.gz"
	println("  --- Downloading")
	COM.DownloadFile(url, homeDir, filename, true, false)
	println("  --- Extracting")
	COM.Extract(homeDir, filename)
	defer os.RemoveAll(filepath.Join(homeDir, strings.TrimSuffix(filename, ".tar.gz")))
	defer COM.CleanupExtract(homeDir, filename)
	pluginDir := filepath.Join(homeDir, "dcevm")
	println("  --- Copying")
	os.MkdirAll(filepath.Join(pluginDir, "bin", "dcevm"), 0755)
	COM.RunCMD("xcopy /E /I /Y \""+filepath.Join(homeDir, strings.TrimSuffix(filename, ".tar.gz"))+"\" \""+pluginDir+"\\\"", false)
	COM.RunCMD("xcopy /E /I /Y \""+filepath.Join(homeDir, strings.TrimSuffix(filename, ".tar.gz"), "bin", "server")+"\" \""+filepath.Join(pluginDir, "bin", "dcevm")+"\\\"", false)
	println("\033[32m  --- DCEVM setup done\033[0m")
}
