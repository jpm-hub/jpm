package doctor

import (
	"fmt"
	COM "jpm/common"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var asked bool = false

func Doctor(silent bool, ask bool) bool {
	asked = ask
	fix := false
	if len(os.Args) > 2 && os.Args[2] == "-fix" {
		fix = true
	}
	Checkverbose()
	good1 := CheckJava(silent)
	fixjava(!good1 && fix)
	good2 := CheckJPX(silent)
	fixjpx(!good2 && fix)
	good3 := CheckGit(silent)
	fixgit(!good3 && fix)
	good4 := checkHotSwapAgent(silent)
	fixhotswap(!good4 && fix)
	good5 := checkJUnit(silent)
	fixjunit(!good5 && fix)
	good6 := checkkotlin(silent)
	fixkotlin(!good6 && fix)
	good7 := checkJar(silent)
	fixjar(!good7 && fix)
	goodFinal := good1 && good2 && good3 && good4 && good5 && good6 && good7
	if goodFinal && asked {
		println("\n\033[32mAll good!\033[0m")
	} else if asked && !fix {
		println("\n\033[32m Fix these issues: jpm doctor -fix \033[0m")
	}
	if fix && asked && !goodFinal {
		os.Args[2] = ""
		Doctor(silent, false)
		println("\n\033[32m You might need to restart your terminal or IDE \033[0m")
	}
	return goodFinal
}

func fixjar(b bool) {
	if !b {
		return
	}
	if !COM.IsWindows() {
		runScript("sdk install openjdk", true)
	} else {
		runScript("winget install Microsoft.OpenJDK.25", true)
	}
}

func fixkotlin(b bool) {
	if !b {
		return
	}
	runScript("jpm setup -kotlin", true)
}

func fixjunit(b bool) {
	if !b {
		return
	}
	runScript("jpm setup -junit", true)
}

func fixhotswap(b bool) {
	if !b {
		return
	}
	runScript("jpm setup -HotSwapAgent", true)
}

func fixgit(b bool) {
	if !b {
		return
	}
	if strings.Contains(runtime.GOOS, "darwin") {
		runScript("brew install git", true)
	} else if strings.Contains(runtime.GOOS, "windows") {
		runScript("winget install Git.Git", true)
	} else {
		runScript("apt-get install git || pacman -S git || dnf install git", true)
	}
}

func fixjpx(b bool) {
	if !b {
		return
	}
	runScript("jpm setup -jpx", true)
}

func fixjava(b bool) {
	if !b {
		return
	}
	runScript("jpm setup -java", true)
}

func Checkverbose() {
	if COM.Verbose {
		fmt.Println("\033[33m[ verbose ]\033[0m", COM.Verbose)
	} else {
		fmt.Println("\033[32m[ verbose ]\033[0m", COM.Verbose)
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
		return false
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
		return false
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
		return false
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
			return false
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
			println("\tfix mac    : brew install git")
			println("\tfix windows: winget install Git.Git")
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
			println("\tfix    : jpm setup -jpx")
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
				println("\tfix windows: winget install Microsoft.OpenJDK.25\n")
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
	if s {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}
	return cmd.Run()
}
