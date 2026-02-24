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
	MakeTestDirsIfNotExists()
	argsMap := COM.ParseArgs()
	os.RemoveAll(filepath.Join(COM.OutDir(), "tests"))
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
		err = COM.RunCMD(COM.ParseEnvVars(prefix, false)+"cd "+COM.OutDir()+" && java "+modular+" -cp \"./;../jpm_dependencies/"+hasDeps+";../jpm_dependencies/tests/*\" org.junit.platform.console.ConsoleLauncher --disable-banner --fail-if-no-tests "+args+extraArgs, true)
	} else {
		if isModular := COM.GetSection("modular", true).(bool); isModular {
			modular = "-p ../jpm_dependencies:../jpm_dependencies/tests"
		}
		err = COM.RunScript(COM.ParseEnvVars(prefix, true)+"cd "+COM.OutDir()+" && java "+modular+" -cp \".:../jpm_dependencies/*:../jpm_dependencies/tests/*\" org.junit.platform.console.ConsoleLauncher --disable-banner --fail-if-no-tests "+args+extraArgs, true)
	}

	if err != nil {
		println("\033[31mAll tests did not succeed\033[0m")
		return err
	}
	println("\033[32mAll pass\033[0m")
	return nil
}

func MakeTestDirsIfNotExists() {
	if _, err := os.Stat("tests"); os.IsNotExist(err) {
		os.Mkdir("tests", os.ModePerm)
		// create a sample test file if not exists

		if strings.Contains(COM.GetSection("language", false).(string), "kotlin") {
			if matches, _ := filepath.Glob(filepath.Join("tests", "*.kt")); len(matches) == 0 {
				sample := strings.ReplaceAll(COM.GetKotlinTestTemplate("", ""), "run()", "\"Hello, World\"")
				sample = strings.ReplaceAll(sample, " Test ", " TestKotlin ")
				os.WriteFile(filepath.Join("tests", "TestKotlin.kt"), []byte(sample), 0644)
			}
		}
		if strings.Contains(COM.GetSection("language", false).(string), "java") {
			if matches, _ := filepath.Glob(filepath.Join("tests", "*.java")); len(matches) == 0 {
				sample := strings.ReplaceAll(COM.GetJavaTestTemplate("", ""), "app.run()", "\"Hello, World\"")
				sample = strings.ReplaceAll(sample, "app = new ();", "")
				sample = strings.ReplaceAll(sample, " Test ", " TestJava ")
				os.WriteFile(filepath.Join("tests", "TestJava.java"), []byte(sample), 0644)
			}
		}
	}
	if _, err := os.Stat(filepath.Join("jpm_dependencies", "tests", "junit.jar")); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join("jpm_dependencies", "tests"), os.ModePerm)
		junitPath := filepath.Join("jpm_dependencies", "tests", "junit.jar")
		if _, err := os.Stat(junitPath); os.IsNotExist(err) {
			os.Link(filepath.Join(COM.HomeDir(), "libs", "junit.jar"), junitPath)
			if strings.Contains(COM.GetSection("language", false).(string), "kotlin") {
				os.Link(filepath.Join(COM.HomeDir(), "libs", "kotlin-test.jar"), filepath.Join("jpm_dependencies", "tests", "kotlin-test.jar"))
			}
		}
	}
}
