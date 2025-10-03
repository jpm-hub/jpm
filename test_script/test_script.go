package test_script

import (
	COM "jpm/common"
	COMPILE "jpm/compile"
	"os"
	"path/filepath"
	"strings"
)

func TestScript() error {
	COM.FindPackageYML(true)
	argsMap := COM.ParseArgs()
	os.RemoveAll(filepath.Join("out", "tests"))
	allTestArgs, found := argsMap["junit"]
	args := ""
	if found {
		args = allTestArgs
	}
	err := COMPILE.CompileTest()
	var err2 error
	if COM.IsWindows() {
		hasDeps := ""
		jpm_dependenciesFiles, _ := os.ReadDir("jpm_dependencies")
		for _, v := range jpm_dependenciesFiles {
			if strings.HasSuffix(v.Name(), ".jar") {
				hasDeps = "*"
				break
			}
		}
		err2 = COM.RunCMD("cd out && "+COM.JAVA()+" -p ../jpm_dependencies;../jpm_dependencies/tests -cp \"./;../jpm_dependencies/"+hasDeps+";../jpm_dependencies/tests/*\" org.junit.platform.console.ConsoleLauncher --class-path tests --scan-classpath --disable-banner --fail-if-no-tests "+args, true)
	} else {
		err2 = COM.RunScript("cd out && "+COM.JAVA()+" -p ../jpm_dependencies:../jpm_dependencies/tests -cp \".:../jpm_dependencies/*:../jpm_dependencies/tests/*\" org.junit.platform.console.ConsoleLauncher --class-path tests --scan-classpath --disable-banner --fail-if-no-tests "+args, true)
	}
	if err != nil {
		return err
	}
	if err2 != nil {
		println("\033[31mAll tests did not succeed\033[0m")
		return err
	}
	println("\033[32mAll pass\033[0m")
	return nil
}
