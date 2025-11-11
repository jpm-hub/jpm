package linux_amd64

import (
	COM "jpm/common"
	"jpm/setup/osless"
	"os"
	"path/filepath"
	"strings"
)

func SetupLinuxAmd64(setupSwitch string, homeDir string) {
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
	// DCEVM for Linux AMD64
	url := "https://cache-redirector.jetbrains.com/intellij-jbr/jbr-25-linux-x64-b176.4.tar.gz"
	filename := "jbr-25-linux-x64-b176.4.tar.gz"
	println("  --- Downloading")
	COM.DownloadFile(url, homeDir, filename, true, false)
	println("  --- Extracting")
	COM.Extract(homeDir, filename)
	defer os.RemoveAll(filepath.Join(homeDir, strings.TrimSuffix(filename, ".tar.gz")))
	defer COM.CleanupExtract(homeDir, filename)
	pluginDir := filepath.Join(homeDir, "dcevm")
	os.MkdirAll(filepath.Join(pluginDir, "libs", "dcevm"), 0755)
	COM.RunScript("cp -r "+filepath.Join(homeDir, strings.TrimSuffix(filename, ".tar.gz"), "*")+" "+pluginDir, false)
	COM.RunScript("cp -r "+filepath.Join(homeDir, strings.TrimSuffix(filename, ".tar.gz"), "lib", "*")+" "+filepath.Join(pluginDir, "libs", "dcevm"), false)
	println("\033[32m  --- DCEVM setup done\033[0m")
}
