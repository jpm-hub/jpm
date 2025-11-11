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
		if !COM.IsWindows() {
			err := runScript(`[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ]`, false)
			if err != nil {
				println(" SDKMAN! is very a lightweight tool to install JVMs, SDKs, JDKs and more.")
				err = runScript("read -p \"Press enter to install sdkman... (ctrl+c to cancel)\n\" ok", true)
				if err == nil {
					runScript("which curl || echo please install curl on you system", true)
					runScript("which unzip || echo please install unzip on you system", true)
					runScript("which zip || echo please install zip on you system", true)
					runScript("which bash || echo please install bash on you system", true)
					err = runScript("which bash&& which zip && which unzip && which curl", true)
					if err != nil {
						println(" Aborting sdkman installation")
						os.Exit(1)
					}
					runScript("curl -s \"https://get.sdkman.io\" | bash", true)
				} else {
					os.Exit(1)
				}
			}
		}
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
	good7 := checkJavac(silent)
	fixjavac(!good7 && fix)
	goodFinal := good1 && good2 && good3 && good4 && good5 && good6 && good7
	if goodFinal && asked {
		println("\n\033[32mAll good!\033[0m")
	} else if asked && !fix {
		println("\n\033[32m Fix these issues: jpm doctor -fix \033[0m")
	}
	if fix && asked && !goodFinal {
		os.Args[2] = ""
		println("\n\033[32m You might need to restart your terminal or IDE \033[0m")
	}
	return goodFinal
}

func fixjavac(b bool) {
	if !b {
		return
	}
	if !COM.IsWindows() {
		runScript("export SDKMAN_DIR=\"$HOME/.sdkman\";[ -s \"$HOME/.sdkman/bin/sdkman-init.sh\" ] && . \"$HOME/.sdkman/bin/sdkman-init.sh\";sdk install java 25.0.1-tem", true)
	} else {
		runScript("winget install Microsoft.OpenJDK.25", true)
	}
}

func fixkotlin(b bool) {
	if !b {
		return
	}
	if !COM.IsWindows() {
		runScript("export SDKMAN_DIR=\"$HOME/.sdkman\";[ -s \"$HOME/.sdkman/bin/sdkman-init.sh\" ] && . \"$HOME/.sdkman/bin/sdkman-init.sh\";sdk install kotlin", true)
	} else {
		runScript("jpm setup -kotlin", true)
	}
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
	runScript("jpm setup -hotswap", true)
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
		runScript("apt-get -y install git || pacman -S -noconfirm git || dnf install -y git || apk add git", true)
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
		fmt.Print("\033[33m[ verbose ] \033[0m", COM.Verbose)
	} else {
		fmt.Print("\033[32m[ verbose ] \033[0m", COM.Verbose)
	}
	if asked {
		println(" -> toggle: 'jpm setup -v'")
	} else {
		println()
	}
}

func checkHotSwapAgent(silent bool) bool {
	hotswapFile := filepath.Join(COM.HomeDir(), "libs", "hotswap-agent.jar")
	if _, err := os.Stat(hotswapFile); err != nil && !silent {
		println("\n\033[33m( hotswap-agent )\033[0m is not accesible (necessary for 'jpm run -hot')")
		if asked {
			println("\tfix: jpm setup -hotswap\n")
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
	if (COM.IsWindows() && COM.KOTLINC() != "kotlinc" || !which("kotlinc")) && !silent {
		println("\n\033[33m( kotlinc )\033[0m is not accesible")
		if asked {
			if !COM.IsWindows() {
				println("\tfix : sdk install kotlin\n")
			} else {
				println("\tfix : jpm setup -kotlin\n")
			}
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
	java := COM.JAVA()
	if java == "java" {
		if !which("java") && !silent {
			println("\n\033[31m( java )\033[0m is not accesible")
			if asked {
				println("\tfix: jpm setup -java  -> uses jetbrains dcevm for 'jpm run -hot'")
			}
			return false
		}
		if !silent {
			println("\n\033[33m( java )\033[0m works with no dcevm")
			if asked {
				println("\tfix: jpm setup -java  -> (recommended) uses jetbrains dcevm for 'jpm run -hot'\n")
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
	if !which("git") && !silent {
		println("\n\033[31m( git )\033[0m is not accesible")
		if asked {
			println("\tfix mac    : brew install git")
			println("\tfix linux  : apt-get install git || pacman -S git || dnf install git || apk add git")
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
	if !which("jpx") && !silent {
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

func checkJavac(silent bool) bool {
	if !which("javac") {
		if !silent {
			println("\n\033[31m( javac )\033[0m is not accesible, you won't be able to compile your app")
			if asked {
				println("\tfix unix   : sdk install java 25.0.1-tem\n")
				println("\tfix windows: winget install Microsoft.OpenJDK.25\n")
			}
		}
		return false
	} else {
		if !silent {
			println("\033[32m( javac )\033[0m works")
		}
	}
	return true
}
func which(cmd string) bool {
	err := runScript("which "+cmd+" || where "+cmd, false)
	return err == nil
}
func runScript(script string, s bool) error {
	var cmd *exec.Cmd
	if COM.IsWindows() {
		cmd = exec.Command("cmd", "/C", script)
	} else {
		err := exec.Command("sh", "-c", "which bash").Run()
		if err == nil {
			cmd = exec.Command("bash", "-c", script)
		} else {
			println("bash shell is not available !!!")
			cmd = exec.Command("sh", "-c", script)
		}
	}
	if s {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}
	return cmd.Run()
}
