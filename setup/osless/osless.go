package osless

import (
	"encoding/json"
	COM "jpm/common"
	JPX "jpm/jpx"
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

func Verbose(homeDir string) {
	configPath := filepath.Join(homeDir, "config.json")

	cfg := make(map[string]interface{})

	// Try to read existing config
	file, err := os.Open(configPath)
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		_ = decoder.Decode(&cfg)
	}
	// Toggle verbose
	currentVerbose, ok := cfg["verbose"].(bool)
	cfg["verbose"] = !ok || !currentVerbose

	// Write config back
	f, err := os.Create(configPath)
	if err != nil {
		println("Failed to write config:", err.Error())
		return
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		println("Failed to encode config:", err.Error())
		return
	}
	println("Verbose set to", cfg["verbose"].(bool))
}

func Jpx() {
	execDir, err := os.Executable()
	if err != nil {
		println("Failed to get executable directory:", err.Error())
		return
	}
	execDir = filepath.Dir(execDir)
	if COM.IsWindows() {
		os.WriteFile(execDir+"\\jpx.cmd", []byte(JPX.SHTemplate()), 0755)
	} else {
		os.WriteFile(execDir+"/jpx", []byte(JPX.SHTemplate()), 0755)
	}
	println("\033[32m  --- jpx setup done\033[0m")
}
