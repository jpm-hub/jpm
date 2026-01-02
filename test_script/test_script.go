package test_script

import (
	COM "jpm/common"
	COMPILE "jpm/compile"
	"os"
	"path/filepath"
	"strings"
)

func TestScript() error {
	extraArgs := strings.TrimSpace(strings.Join(os.Args[2:], " "))
	if len(extraArgs) == 0 {
		extraArgs = "--scan-classpath"
	}
	extraArgs = " " + extraArgs
	COM.FindPackageYML(true)
	argsMap := COM.ParseArgs()
	os.RemoveAll(filepath.Join("out", "tests"))
	modular := ""
	allTestArgs, found := argsMap["junit"]
	args := ""
	if found {
		args = allTestArgs
	}
	err := COMPILE.CompileTest()
	if err != nil {
		return err
	}
	prefix := "export "
	if COM.IsWindows() {
		prefix = "set "
		hasDeps := ""
		jpm_dependenciesFiles, _ := os.ReadDir("jpm_dependencies")
		for _, v := range jpm_dependenciesFiles {
			if strings.HasSuffix(v.Name(), ".jar") {
				hasDeps = "*"
				break
			}
		}
		if isModular := COM.GetSection("modular", true).(bool); isModular {
			modular = "-p ../jpm_dependencies;../jpm_dependencies/tests"
		}
		err = COM.RunCMD(COM.ParseEnvVars(prefix, false)+"cd out && java "+modular+" -cp \"./;../jpm_dependencies/"+hasDeps+";../jpm_dependencies/tests/*\" org.junit.platform.console.ConsoleLauncher --disable-banner --fail-if-no-tests "+args+extraArgs, true)
	} else {
		if isModular := COM.GetSection("modular", true).(bool); isModular {
			modular = "-p ../jpm_dependencies:../jpm_dependencies/tests"
		}
		err = COM.RunScript(COM.ParseEnvVars(prefix, true)+"cd out && java "+modular+" -cp \".:../jpm_dependencies/*:../jpm_dependencies/tests/*\" org.junit.platform.console.ConsoleLauncher --disable-banner --fail-if-no-tests "+args+extraArgs, true)
	}

	if err != nil {
		println("\033[31mAll tests did not succeed\033[0m")
		return err
	}
	println("\033[32mAll pass\033[0m")
	return nil
}
