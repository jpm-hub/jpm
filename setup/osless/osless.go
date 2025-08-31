package osless

import (
	COM "jpm/common"
	"os"
	"path/filepath"
)

func Kotlinc(homeDir string) {
	url := "https://github.com/JetBrains/kotlin/releases/download/v2.2.0/kotlin-compiler-2.2.0.zip"
	filename := "kotlin-compiler-2.2.0.zip"
	println("  --- Downloading")
	COM.DownloadFile(url, homeDir, filename, true, false)
	println("  --- Extracting")
	COM.Extract(homeDir, filename)
	defer COM.CleanupExtract(homeDir, filename)
	println("\033[32m  --- Kolinc setup done\033[0m")
}
func JunitPlugin(homeDir string) {
	url := "https://repo1.maven.org/maven2/org/junit/platform/junit-platform-console-standalone/1.9.3/junit-platform-console-standalone-1.9.3.jar"
	filename := "junit-platform-console-standalone-1.9.3.jar"
	pluginDir := filepath.Join(homeDir, "libs")
	os.MkdirAll(pluginDir, 0755)
	println("  --- Downloading")
	COM.DownloadFile(url, pluginDir, filename, true, false)
	os.Rename(filepath.Join(pluginDir, filename), filepath.Join(pluginDir, "junit.jar"))
	println("\033[32m  --- JUnit5 setup done\033[0m")
}
func HotSwapAgentPlugin(homeDir string) {
	url := "https://github.com/HotswapProjects/HotswapAgent/releases/download/RELEASE-2.0.0/hotswap-agent-2.0.0.jar"
	filename := "hotswap-agent-2.0.0.jar"
	pluginDir := filepath.Join(homeDir, "libs")
	os.MkdirAll(pluginDir, 0755)
	println("  --- Downloading")
	COM.DownloadFile(url, pluginDir, filename, true, false)
	os.Rename(filepath.Join(pluginDir, filename), filepath.Join(pluginDir, "hotswap-agent.jar"))
	println("\033[32m  --- Hotswap-Agent setup done\033[0m")
}
