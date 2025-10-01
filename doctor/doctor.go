package doctor

import (
	"fmt"
	COM "jpm/common"
	"os"
	"os/exec"
	"path/filepath"
)

var asked bool = false

func Doctor(silent bool, ask bool) bool {
	asked = ask
	CheckVerobse()
	good := CheckJava(silent)
	good = CheckJPX(silent) && good
	good = CheckGit(silent) && good
	good = checkHotSwapAgent(silent) && good
	good = checkJUnit(silent) && good
	good = checkkotlin(silent) && good
	good = checkJar(silent) && good
	return good
}

func CheckVerobse() {
	if COM.Verbose {
		fmt.Println("\033[33m[ Verobse ]\033[0m", COM.Verbose)
	} else {
		fmt.Println("\033[32m[ Verobse ]\033[0m", COM.Verbose)
	}
	if asked {
		println("\ttoggle: jpm setup -v")
	}
}

func checkHotSwapAgent(silent bool) bool {
	hotswapFile := filepath.Join(COM.HomeDir(), "libs", "hotswap-agent.jar")
	if _, err := os.Stat(hotswapFile); err != nil && !silent {
		println("\n\033[33m( hotswap-agent )\033[0m is not accesible (necessary for 'jpm run -hot')")
		if asked {
			println("\tfix: jpm setup -HotSwapAgent\n")
		}
	} else {
		if !silent {
			println("\033[32m( hotswap-agent )\033[0m works")
		}
	}
	return true
}
func checkJUnit(silent bool) bool {
	hotswapFile := filepath.Join(COM.HomeDir(), "libs", "junit.jar")
	if _, err := os.Stat(hotswapFile); err != nil && !silent {
		println("\n\033[33m( junit5 )\033[0m is not accesible (necessary for 'jpm test' with no dependecies)")
		if asked {
			println("\tfix: jpm setup -junit\n")
		}
	} else {
		if !silent {
			println("\033[32m( junit5 )\033[0m works")
		}
	}
	return true
}
func checkkotlin(silent bool) bool {
	if COM.KOTLINC() == "" && !silent {
		println("\n\033[33m( kotlinc )\033[0m is not accesible")
		if asked {
			println("\tfix: jpm setup -kotlin\n")
		}
	} else {
		if !silent {
			println("\033[32m( kotlin )\033[0m works")
		}
	}
	return true
}
func CheckJava(silent bool) bool {
	javac := COM.JAVAC()
	if javac == "javac" {
		if err := runScript("which javac || where javac", false); err != nil && !silent {
			println("\n\033[31m( java )\033[0m is not accesible")
			if asked {
				println("\tfix: jpm setup -java   -> uses jetbrains dcevm for 'jpm run -hot'")
			}
			return false
		}
		if !silent {
			println("\n\033[33m( java )\033[0m works with no dcevm")
			if asked {
				println("\tfix: jpm setup -java   -> (recommended) uses jetbrains dcevm for 'jpm run -hot'\n")
			}
			return true
		}
		return false
	} else {
		if !silent {
			println("\033[32m( java )\033[0m works")
		}
	}
	return true
}

func CheckGit(silent bool) bool {
	if err := runScript("which git || where git", false); err != nil && !silent {
		println("\n\033[31m( git )\033[0m is not accesible")
		if asked {
			println("\tfix mac    : brew install git'")
			println("\tfix windows: winget install git'")
		}
		return false
	} else {
		if !silent {
			println("\033[32m( git )\033[0m works")
		}
	}
	return true
}

func CheckJPX(silent bool) bool {
	if err := runScript("which jpx || where jpx", false); err != nil && !silent {
		println("\n\033[31m( jpx )\033[0m is not accesible")
		if asked {
			println("\tfix    : jpm setup -jpx'")
		}
		return false
	} else {
		if !silent {
			println("\033[32m( jpx )\033[0m works")
		}
	}
	return true
}

func checkJar(silent bool) bool {
	if err := runScript("which jar || where jar", false); err != nil {
		if !silent {
			println("\n\033[31m( jar )\033[0m is not accesible, you won't be able to bundle your app")
			if asked {
				println("\tfix mac    : sdk install openjdk\n")
				println("\tfix windows: winget install openjdk\n")
			}
		}
		return false
	} else {
		if !silent {
			println("\033[32m( jar )\033[0m works")
		}
	}
	return true
}

func runScript(script string, s bool) error {
	var cmd *exec.Cmd
	if COM.IsWindows() {
		// On Windows, use cmd.exe to create a process group
		cmd = exec.Command("cmd", "/C", script)
	} else {
		// On Unix, use sh -c with exec to ensure proper process group
		cmd = exec.Command("sh", "-c", script)
	}
	return cmd.Run()
}
