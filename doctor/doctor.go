package doctor

import (
	COM "jpm/common"
	"os"
	"path/filepath"
)

func Doctor(silent bool) bool {
	good := CheckJava(silent)

	good = checkHotSwapAgent(silent) && good
	good = checkJUnit(silent) && good
	good = checkkotlin(silent) && good
	good = checkJar(silent) && good
	return good
}

//	func checkCommand(cmd string, help string, silent bool) bool {
//		if err := exec.Command("sh", "-c", "which "+cmd).Run(); err != nil && !silent {
//			println("\n\033[31m( " + cmd + " )\033[0m is not accesible")
//			println("\t" + help + "\n")
//			return false
//		} else {
//			if !silent {
//				println("\033[32m( " + cmd + " )\033[0m works")
//			}
//			return true
//		}
//	}

func checkHotSwapAgent(silent bool) bool {
	hotswapFile := filepath.Join(COM.HomeDir(), "libs", "hotswap-agent.jar")
	if _, err := os.Stat(hotswapFile); err != nil && !silent {
		println("\n\033[33m( hotswap-agent )\033[0m is not accesible (necessary for 'jpm run -hot')")
		println("\tfix: jpm setup -HotSwapAgent\n")
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
		println("\tfix: jpm setup -junit\n")
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
		println("\tfix: jpm setup -kotlin\n")
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
		if err := COM.RunScript("which javac || where javac", false); err != nil && !silent {
			println("\n\033[31m( java )\033[0m is not accesible")
			println("\tfix: jpm setup -java   -> uses jetbrains dcevm for 'jpm run -hot'")
			println("\tor: brew install openjdk\n")
			return false
		}
		if !silent {
			println("\n\033[33m( java )\033[0m works with no dcevm")
			println("\tfix: jpm setup -java   -> (recommended) uses jetbrains dcevm for 'jpm run -hot'\n")
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

func checkJar(silent bool) bool {
	if err := COM.RunScript("which jar || where jar", false); err != nil {
		if !silent {
			println("\n\033[31m( jar )\033[0m is not accesible, you won't be able to bundle your app")
			println("\ttfix: brew install openjdk\n")
		}
		return false
	} else {
		if !silent {
			println("\033[32m( jar )\033[0m works")
		}
	}
	return true
}
