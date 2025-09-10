package compile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	COM "jpm/common"
)

func compileTestJava() error {
	args := ""
	if allBuildArgs, found := argsMap["javac"]; found {
		args = allBuildArgs
	}
	os.MkdirAll(filepath.Join("out", "tests"), 0755)
	jpm_dependenciesFiles, err := os.ReadDir("jpm_dependencies")
	jpm_dependenciesFiles2, err2 := os.ReadDir(filepath.Join("jpm_dependencies", "tests"))
	if err != nil || err2 != nil {
		return fmt.Errorf("failed to read ./jpm_dependencies/tests")
	}
	var builder strings.Builder
	for _, file := range jpm_dependenciesFiles2 {
		if strings.HasSuffix(file.Name(), ".jar") {
			if builder.Len() > 0 {
				builder.WriteString(separator)
			}
			builder.WriteString(filepath.Join("jpm_dependencies", "tests") + "/")
			builder.WriteString(file.Name())
		}
	}
	for _, file := range jpm_dependenciesFiles {
		if strings.HasSuffix(file.Name(), ".jar") || strings.HasSuffix(file.Name(), ".zip") {
			if builder.Len() > 0 {
				builder.WriteString(separator)
			}
			builder.WriteString("jpm_dependencies/")
			builder.WriteString(file.Name())
		}
	}
	jarFilesString := builder.String()
	r, w, originalOut, err := startCheckLastLineForErrors()
	if err != nil {
		return err
	}
	allJavas := strings.Join([]string{findAllSrcFile(COM.SrcDir(), "*.java"), findAllSrcFile("tests", "*.java")}, " ")
	var err4 error
	if COM.IsWindows() {
		err4 = COM.RunCMD(COM.JAVAC()+" -proc:full "+args+" -cp \""+"out/tests;"+jarFilesString+"\" -d out/tests "+allJavas, true)
	} else {
		err4 = COM.RunScript(COM.JAVAC()+" -proc:full "+args+" -cp \""+"out/tests:"+jarFilesString+"\" -d out/tests "+allJavas, true)
	}

	err5 := endCheckLastLineForErrors(r, w, originalOut)
	if err4 != nil || err5 != nil {
		return fmt.Errorf("\033[31m test compilation failed for java\033[0m")
	}
	return nil
}
