package test_script

import (
	COM "jpm/common"
	COMPILE "jpm/compile"
	"os"
	"path/filepath"
)

func TestScript() error {
	COM.FindPackageYML()
	argsMap := COM.ParseArgs()
	os.RemoveAll(filepath.Join("out", "tests"))
	allTestArgs, found := argsMap["test"]
	args := ""
	if found {
		args = allTestArgs
	}
	err := COMPILE.CompileTest()
	var err2 error
	if COM.IsWindows() {
		err2 = COM.RunCMD("cd out && "+COM.JAVA()+" -cp \"./tests;../jpm_dependencies/*;../jpm_dependencies/tests/*\" org.junit.platform.console.ConsoleLauncher  --class-path tests --scan-classpath --disable-banner "+args, true)
	} else {
		err2 = COM.RunScript("cd out && "+COM.JAVA()+" -cp \"../jpm_dependencies/*:../jpm_dependencies/tests/*\" org.junit.platform.console.ConsoleLauncher  --class-path tests --scan-classpath --disable-banner "+args, true)
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
